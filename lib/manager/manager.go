package manager

import (
	"fmt"
	"time"

	"github.com/r3boot/go-musicbot/lib/utils"

	"github.com/r3boot/go-musicbot/lib/liquidsoap"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/dbclient"
	"github.com/r3boot/go-musicbot/lib/log"
	"github.com/r3boot/go-musicbot/lib/tags"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

const (
	maxAddTracksJobs = 32
	maxPriority      = 10
)

type Manager struct {
	cfg *config.Config
	ls  *liquidsoap.LSClient
	db  *dbclient.DbClient
	yt  *ytclient.YoutubeClient
}

func NewManager(cfg *config.Config) (*Manager, error) {
	var err error

	mgr := &Manager{
		cfg: cfg,
	}

	if cfg == nil {
		log.Fatalf(log.Fields{
			"package":  "manager",
			"function": "NewManager",
		}, "unable to initialize: cfg == nil")
	}

	mgr.ls, err = liquidsoap.NewLSClient(cfg.Liquidsoap)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "manager",
			"function": "NewManager",
			"call":     "liquidsoap.NewLSClient",
		}, err.Error())
	}

	mgr.db, err = dbclient.NewDbClient(cfg.Postgres)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "manager",
			"function": "NewManager",
			"call":     "dbclient.NewDbClient",
		}, err.Error())
	}

	mgr.yt, err = ytclient.NewYoutubeClient(cfg)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "manager",
			"function": "NewManager",
			"call":     "ytclient.NewYoutubeClient",
		}, err.Error())
	}

	_, err = mgr.Synchronize()
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "manager",
			"function": "NewManager",
			"call":     "mgr.Synchronize",
		}, err.Error())
	}

	return mgr, nil
}

func (m *Manager) IsAllowedLength(yid string) bool {
	err := m.yt.IsAllowedLength(yid)
	if err != nil {
		return false
	}
	return true
}

func (m *Manager) HasYid(yid string) bool {
	_, err := m.db.GetTrackByYid(yid)
	if err == nil {
		return true
	}
	return false
}

func (m *Manager) AddTrack(yid, submitter string) (*dbclient.Track, error) {
	fname, err := m.yt.Download(&ytclient.DownloadJob{
		Yid:       yid,
		Submitter: submitter,
	})
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "manager",
			"function":  "AddTrack",
			"call":      "m.yt.Download",
			"yid":       yid,
			"submitter": submitter,
		}, err.Error())
		return nil, fmt.Errorf("yt.Download: %v", err)
	}

	duration, err := tags.GetDuration(m.cfg.Datastore.Directory + "/" + fname)
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "manager",
			"function":  "AddTrack",
			"call":      "tags.GetDuration",
			"yid":       yid,
			"submitter": submitter,
		}, err.Error())
		return nil, fmt.Errorf("tags.GetDuration: %v", err)
	}

	track := &dbclient.Track{
		Filename:  fname,
		Yid:       yid,
		Rating:    5,
		Submitter: submitter,
		Duration:  duration,
		AddedOn:   time.Now(),
	}

	err = track.Add()
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "manager",
			"function":  "AddTrack",
			"call":      "track.Add",
			"yid":       yid,
			"submitter": submitter,
		}, err.Error())
		return nil, fmt.Errorf("track.Add: %v", err)
	}

	return track, nil
}

func (m *Manager) Search(query, submitter string) ([]dbclient.Track, error) {
	tracks, err := m.db.Search(query)
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "manager",
			"function":  "Search",
			"call":      "m.db.Search",
			"query":     query,
			"submitter": submitter,
		}, err.Error())
		return nil, fmt.Errorf("db.Search: %v", err)
	}

	return tracks, nil
}

func (m *Manager) Request(query, submitter string) (dbclient.Track, error) {
	tracks, err := m.db.Search(query)
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "manager",
			"function":  "Request",
			"call":      "m.db.Search",
			"query":     query,
			"submitter": submitter,
		}, err.Error())
		return dbclient.Track{}, fmt.Errorf("Search: %v", err)
	}
	if len(tracks) == 0 {
		return dbclient.Track{}, fmt.Errorf("No results found")
	}

	track := tracks[0]
	entry := liquidsoap.PlaylistEntry{
		Filename: m.cfg.Datastore.Directory + "/" + track.Filename,
	}
	m.ls.Enqueue(&entry)

	log.Infof(log.Fields{
		"package":   "manager",
		"function":  "Request",
		"filename":  track.Filename,
		"query":     query,
		"submitter": submitter,
	}, "track added to queue")

	return track, nil
}

func (m *Manager) IncreaseRating() error {
	track, err := m.NowPlaying()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "IncreaseRating",
			"call":     "m.NowPlaying",
		}, err.Error())
		return fmt.Errorf("NowPlaying: %v", err)
	}

	track.Rating += 1
	err = track.Save()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "IncreaseRating",
			"call":     "track.Save",
		}, err.Error())
		return fmt.Errorf("failed to increase rating")
	}

	log.Infof(log.Fields{
		"package":  "manager",
		"function": "IncreaseRating",
		"filename": track.Filename,
		"rating":   track.Rating,
	}, "rating increased")

	return nil
}

func (m *Manager) DecreaseRating() error {
	track, err := m.NowPlaying()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "DecreaseRating",
			"call":     "m.NowPlaying",
		}, err.Error())
		return fmt.Errorf("NowPlaying: %v", err)
	}

	track.Rating -= 1
	err = track.Save()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "IncreaseRating",
			"call":     "track.Save",
		}, err.Error())
		return fmt.Errorf("failed to increase rating")
	}

	// TODO: Handle track delete on rating == 0

	log.Infof(log.Fields{
		"package":  "manager",
		"function": "DecreaseRating",
		"filename": track.Filename,
		"rating":   track.Rating,
	}, "rating decreased")

	return nil
}

func (m *Manager) Next() error {
	err := m.ls.Next()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "Next",
			"call":     "m.ls.Next",
		}, err.Error())
	}

	return err
}

func (m *Manager) NowPlaying() (*dbclient.Track, error) {
	nowPlaying, err := m.ls.NowPlaying()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "NowPlaying",
			"call":     "m.ls.NowPlaying",
		}, err.Error())
		return nil, fmt.Errorf("NowPlaying: %v", err)
	}

	yid, err := utils.GetYidFromFilename(nowPlaying.Filename)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "NowPlaying",
			"call":     "utils.GetYidFromFilename",
			"filename": nowPlaying.Filename,
		}, err.Error())
		return nil, fmt.Errorf("GetYidFromFilename: %v", err)
	}

	track, err := m.db.GetTrackByYid(yid)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "NowPlaying",
			"call":     "m.db.GetTrackByYid",
			"yid":      yid,
		}, err.Error())
		return nil, fmt.Errorf("GetTrackByYid: %v", err)
	}
	track.Elapsed = nowPlaying.Elapsed

	return track, nil
}

func (m *Manager) GetQueue() ([]*dbclient.Track, error) {
	nowPlaying, err := m.NowPlaying()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "GetQueue",
			"call":     "m.NowPlaying",
		}, err.Error())
		return nil, fmt.Errorf("failed to fetch nowplaying info")
	}

	queueEntries := []*dbclient.Track{}

	foundEntries, err := m.ls.GetQueue()
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "manager",
			"function": "GetQueue",
			"call":     "m.ls.GetQueue",
		}, err.Error())
		return nil, fmt.Errorf("GetQueue: %v", err)
	}

	for _, foundEntry := range foundEntries {
		if foundEntry.Filename == nowPlaying.Filename {
			continue
		}

		yid, err := utils.GetYidFromFilename(foundEntry.Filename)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "manager",
				"function": "GetQueue",
				"call":     "utils.GetYidFromFilename",
				"filename": foundEntry.Filename,
			}, err.Error())
			continue
		}
		track, err := m.db.GetTrackByYid(yid)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "manager",
				"function": "GetQueue",
				"call":     "m.db.GetTrackByYid",
				"filename": yid,
			}, err.Error())
			return nil, fmt.Errorf("GetTrackByYid: %v", err)
		}

		queueEntries = append(queueEntries, track)
	}

	return queueEntries, nil
}

func (m *Manager) Synchronize() (int, error) {
	trackPattern := m.cfg.Datastore.Directory + "/*.mp3"
	tracks, err := tags.ReadTagsFrom(trackPattern)
	if err != nil {
		log.Warningf(log.Fields{
			"package":       "manager",
			"function":      "Synchronize",
			"call":          "tags.ReadTagsFrom",
			"track_pattern": trackPattern,
		}, err.Error())
		return -1, fmt.Errorf("ReadTagsFrom: %v", err)
	}

	numSynchronized := 0
	for _, track := range tracks {
		_, err := m.db.GetTrackByYid(track.Yid)
		if err == nil {
			continue
		}

		err = track.Add()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "manager",
				"function":  "Synchronize",
				"call":      "track.Add",
				"filename":  track.Filename,
				"submitter": track.Submitter,
			}, err.Error())
			continue
		}
		numSynchronized += 1
	}

	if numSynchronized > 0 {
		log.Infof(log.Fields{
			"package":   "manager",
			"function":  "Synchronize",
			"num_added": numSynchronized,
		}, "added tracks to the database")
	}

	return numSynchronized, nil
}
