package indexer

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/r3boot/go-musicbot/pkg/config"
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	ModuleName = "Search"
)

type ResultsView struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

type Search struct {
	sendChan chan mq.Message
	recvChan chan mq.Message
	quit     chan bool
	log      *logrus.Entry
	cfg      *config.MusicBotConfig
	idx      bleve.Index
}

func NewSearch(cfg *config.MusicBotConfig) (*Search, error) {
	s := &Search{
		sendChan: make(chan mq.Message, mq.MaxInFlight),
		recvChan: make(chan mq.Message, mq.MaxInFlight),
		quit:     make(chan bool),
		log:      logrus.WithFields(logrus.Fields{"caller": ModuleName}),
		cfg:      cfg,
	}

	err := s.Open()
	if err != nil {
		return nil, fmt.Errorf("s.Open: %v", err)
	}

	go s.MessagePipe()

	return s, nil
}

func (s *Search) Open() error {
	// Check if on-disk store is already created
	hasIndex := false
	fs, err := os.Stat(s.cfg.Paths.Index)
	if err == nil && fs.IsDir() {
		hasIndex = true
	}

	if hasIndex {
		s.idx, err = bleve.Open(s.cfg.Paths.Index)
		if err != nil {
			return fmt.Errorf("bleve.Open: %v", err)
		}
	} else {
		// Setup index
		indexMapping := bleve.NewIndexMapping()

		mapping := bleve.NewDocumentMapping()

		fnameField := bleve.NewTextFieldMapping()
		mapping.AddFieldMappingsAt("Filename", fnameField)

		lastModifiedField := bleve.NewDateTimeFieldMapping()
		mapping.AddFieldMappingsAt("LastModified", lastModifiedField)

		ratingField := bleve.NewNumericFieldMapping()
		mapping.AddFieldMappingsAt("Rating", ratingField)

		durationField := bleve.NewNumericFieldMapping()
		mapping.AddFieldMappingsAt("Duration", durationField)

		posField := bleve.NewNumericFieldMapping()
		mapping.AddFieldMappingsAt("Pos", posField)

		idField := bleve.NewNumericFieldMapping()
		mapping.AddFieldMappingsAt("Id", idField)

		submitterField := bleve.NewTextFieldMapping()
		mapping.AddFieldMappingsAt("Submitter", submitterField)

		indexMapping.AddDocumentMapping("_default", mapping)

		s.idx, err = bleve.New(s.cfg.Paths.Index, indexMapping)
		if err != nil {
			return fmt.Errorf("bleve.New: %v", err)
		}
	}

	return nil
}

func (s *Search) MessagePipe() {
	s.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-s.recvChan:
			{
				msgType := msg.GetMsgType()
				switch msgType {
				case mq.MsgRequest:
					{
						requestMsg := msg.GetContent().(*mq.RequestMessage)
						s.searchRequest(requestMsg.Query, requestMsg.Submitter)
					}
				case mq.MsgUpdateIndex:
					{
						updateIndexMsg := msg.GetContent().(*mq.UpdateIndexMessage)
						s.update(updateIndexMsg.Filename, updateIndexMsg.Pos)
					}
				case mq.MsgUpdatePlaylist:
					{
						upMsg := msg.GetContent().(*mq.UpdatePlaylistMessage)
						s.bulkUpdate(upMsg.Entries)
					}
				}
			}
		case <-s.quit:
			{
				return
			}
		}
	}
}

func (s *Search) GetRecvChan() chan mq.Message {
	return s.recvChan
}

func (s *Search) GetSendChan() chan mq.Message {
	return s.sendChan
}

func (s *Search) searchRequest(q, submitter string) {
	query := bleve.NewMatchQuery(q)
	search := bleve.NewSearchRequest(query)

	srLog := s.log.WithFields(logrus.Fields{
		"q":         q,
		"submitter": submitter,
	})

	results, err := s.idx.Search(search)
	if err != nil {
		e := fmt.Sprintf("s.idx.Search: %v", err)
		srLog.WithFields(logrus.Fields{
			"error": e,
		}).Warn("Search failed")
		errMsg := mq.NewSearchErrorMessage(e, submitter)
		msg := mq.NewMessage(ModuleName, mq.MsgSearchError, errMsg)
		s.sendChan <- *msg
		return
	}

	if len(results.Hits) == 0 {
		e := "no results found"
		srLog.WithFields(logrus.Fields{
			"error": e,
		}).Warn("Search failed")
		errMsg := mq.NewSearchErrorMessage(e, submitter)
		msg := mq.NewMessage(ModuleName, mq.MsgSearchError, errMsg)
		s.sendChan <- *msg
		return
	}

	id := results.Hits[0].ID

	entry, err := s.documentToPlaylistEntry(id)
	if err != nil {
		e := fmt.Sprintf("s.documentToPlaylistEntry: %v", err)
		srLog.WithFields(logrus.Fields{
			"error": e,
		}).Warn("Search failed")
		errMsg := mq.NewSearchErrorMessage(e, submitter)
		msg := mq.NewMessage(ModuleName, mq.MsgSearchError, errMsg)
		s.sendChan <- *msg
		return
	}

	srLog.Debug("Sending Queue request")
	queueMsg := mq.NewQueueMessage(entry.Id, entry.Filename, submitter)
	msg := mq.NewMessage(ModuleName, mq.MsgQueue, queueMsg)
	s.sendChan <- *msg
}

func (s *Search) update(fname string, pos int) {
	s.log.WithFields(logrus.Fields{
		"fname": fname,
		"pos":   pos,
	}).Debug("Adding track to index")
}

func (s *Search) bulkUpdate(playlist map[string]mq.PlaylistEntry) {
	tStart := time.Now()
	numIndexed := 0

	batchJob := s.idx.NewBatch()

	for fname, entry := range playlist {
		if err := batchJob.Index(fname, entry); err != nil {
			s.log.WithFields(logrus.Fields{
				"fname": fname,
				"error": err,
			}).Warn("Failed to update entry")
			continue
		}
		numIndexed += 1
	}

	s.idx.Batch(batchJob)
	tDuration := time.Since(tStart)
	s.log.WithFields(logrus.Fields{
		"duration":   tDuration,
		"numtracks":  len(playlist),
		"numindexed": numIndexed,
	}).Info("Updated index")
}

func (s *Search) documentToPlaylistEntry(id string) (*mq.PlaylistEntry, error) {
	doc, err := s.idx.Document(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve document")
	}

	entry := &mq.PlaylistEntry{}

	for _, field := range doc.Fields {
		var value interface{}
		switch field := field.(type) {
		case *document.TextField:
			{
				value = string(field.Value())
			}
		case *document.NumericField:
			{
				num, err := field.Number()
				if err != nil {
					return nil, fmt.Errorf("failed to convert number")
				}
				value = int(num)
			}
		case *document.DateTimeField:
			{
				value, err = field.DateTime()
				if err != nil {
					return nil, fmt.Errorf("failed to convert datetime")
				}
			}
		}
		switch field.Name() {
		case "Filename":
			{
				entry.Filename = value.(string)
			}
		case "LastModified":
			{
				entry.LastModified = value.(time.Time)
			}
		case "Rating":
			{
				entry.Rating = value.(int)
			}
		case "Duration":
			{
				entry.Duration = value.(int)
			}
		case "Pos":
			{
				entry.Pos = value.(int)
			}
		case "Id":
			{
				entry.Id = value.(int)
			}
		case "Submitter":
			{
				entry.Submitter = value.(string)
			}
		}
	}

	return entry, nil
}

func (s *Search) FindByFilename(fname string) (*mq.PlaylistEntry, error) {
	q := fmt.Sprintf("Filename:%s", fname)
	query := bleve.NewMatchQuery(q)
	search := bleve.NewSearchRequest(query)
	results, err := s.idx.Search(search)
	if err != nil {
		return nil, fmt.Errorf("s.idx.Search: %v", err)
	}

	if len(results.Hits) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	id := results.Hits[0].ID

	if id != fname {
		return nil, fmt.Errorf("filename not found")
	}

	return s.documentToPlaylistEntry(id)
}
