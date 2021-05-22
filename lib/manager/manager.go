package manager

import (
	"fmt"
	"log"
	"time"

	"github.com/r3boot/test/lib/utils"

	"github.com/r3boot/test/lib/liquidsoap"

	"github.com/r3boot/test/lib/config"
	"github.com/r3boot/test/lib/dbclient"
	"github.com/r3boot/test/lib/tags"
	"github.com/r3boot/test/lib/ytclient"
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
	// playQueue *PlayQueue
}

func NewManager(cfg *config.Config) (*Manager, error) {
	var err error

	mgr := &Manager{
		cfg: cfg,
	}

	if cfg == nil {
		return nil, fmt.Errorf("cfg == nil")
	}

	mgr.ls, err = liquidsoap.NewLSClient(cfg.Liquidsoap)
	if err != nil {
		return nil, fmt.Errorf("NewLSClient: %v", err)
	}

	mgr.db, err = dbclient.NewDbClient(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("NewDbClient: %v", err)
	}

	mgr.yt, err = ytclient.NewYoutubeClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("NewYoutubeClient: %v", err)
	}

	// mgr.playQueue = NewPlayQueue(mgr.ls)

	_, err = mgr.Synchronize()
	if err != nil {
		return nil, fmt.Errorf("Synchronize: %v", err)
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
		return nil, fmt.Errorf("yt.Download: %v", err)
	}

	duration, err := tags.GetDuration(m.cfg.Datastore.Directory + "/" + fname)
	if err != nil {
		return nil, fmt.Errorf("GetDuration: %v", err)
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
		return nil, fmt.Errorf("track.Add: %v", err)
	}

	return track, nil
}

func (m *Manager) Search(query, submitter string) ([]dbclient.Track, error) {
	tracks, err := m.db.Search(query)
	if err != nil {
		return nil, fmt.Errorf("Search: %v", err)
	}

	return tracks, nil
}

func (m *Manager) Request(query, submitter string) (dbclient.Track, error) {
	tracks, err := m.db.Search(query)
	if err != nil {
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

	return track, nil
}

func (m *Manager) IncreaseRating() error {
	track, err := m.NowPlaying()
	if err != nil {
		return fmt.Errorf("NowPlaying: %v", err)
	}

	track.Rating += 1
	track.Save()
	return nil
}

func (m *Manager) DecreaseRating() error {
	track, err := m.NowPlaying()
	if err != nil {
		return fmt.Errorf("NowPlaying: %v", err)
	}

	track.Rating -= 1
	track.Save()

	return nil
}

func (m *Manager) Next() error {
	err := m.ls.Next()
	return err
}

func (m *Manager) NowPlaying() (*dbclient.Track, error) {
	nowPlaying, err := m.ls.NowPlaying()
	if err != nil {
		return nil, fmt.Errorf("NowPlaying: %v", err)
	}

	yid, err := utils.GetYidFromFilename(nowPlaying.Filename)
	if err != nil {
		return nil, fmt.Errorf("GetYidFromFilename: %v", err)
	}

	track, err := m.db.GetTrackByYid(yid)
	if err != nil {
		return nil, fmt.Errorf("GetTrackByYid: %v", err)
	}

	return track, nil
}

func (m *Manager) GetQueue() ([]*dbclient.Track, error) {
	queueEntries := []*dbclient.Track{}

	foundEntries, err := m.ls.GetQueue()
	if err != nil {
		return nil, fmt.Errorf("GetQueue: %v", err)
	}

	for _, foundEntry := range foundEntries {
		yid, err := utils.GetYidFromFilename(foundEntry.Filename)
		if err != nil {
			log.Printf("GetYidFromFilename: %v\n", err)
		}
		track, err := m.db.GetTrackByYid(yid)
		if err != nil {
			return nil, fmt.Errorf("GetTrackById: %v", err)
		}

		queueEntries = append(queueEntries, track)
	}

	return queueEntries, nil
}

func (m *Manager) Synchronize() (int, error) {
	trackPattern := m.cfg.Datastore.Directory + "/*.mp3"
	tracks, err := tags.ReadTagsFrom(trackPattern)
	if err != nil {
		return -1, fmt.Errorf("ReadTagsFrom: %v", err)
	}

	numSynchronized := 0
	for _, track := range tracks {
		tmp, err := m.db.GetTrackByYid(track.Yid)
		if err == nil {
			tmp.Yid = track.Yid
			err = tmp.Save()
			if err != nil {
				return -1, fmt.Errorf("track.Update: %v", err)
			}
		} else {
			err = track.Add()
			if err != nil {
				return -1, fmt.Errorf("track.Add: %v", err)
			}
		}
		numSynchronized += 1
	}

	return numSynchronized, nil
}
