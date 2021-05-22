package downloader

import (
	"bytes"
	"fmt"
	"github.com/r3boot/go-musicbot/pkg/config"
	"github.com/r3boot/go-musicbot/pkg/id3tags"
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	ModuleName = "Downloader"

	MaxSongLength = 1800 // In Seconds

	MaxDownloadsInFlight = 64

	SiteYoutube = iota
)

var (
	SiteToString = map[int]string{
		SiteYoutube: "Youtube",
	}
	reIDs = map[int]*regexp.Regexp{
		SiteYoutube: regexp.MustCompile(".*-([a-zA-Z0-9_-]{11}).mp3"),
	}

	reSongLength = regexp.MustCompile("\"length_seconds\":\"([0-9]+)\"")
)

type DownloadRequest struct {
	Site      int
	ID        string
	Submitter string
}

type Download struct {
	sendChan chan mq.Message
	recvChan chan mq.Message
	quit     chan bool
	log      *logrus.Entry
	cfg      *config.MusicBotConfig
	tags     *id3tags.ID3Tags
	jobChan  chan DownloadRequest
	idList   []string
}

func NewDownloader(cfg *config.MusicBotConfig, tags *id3tags.ID3Tags) (*Download, error) {
	dl := &Download{
		sendChan: make(chan mq.Message, mq.MaxInFlight),
		recvChan: make(chan mq.Message, mq.MaxInFlight),
		quit:     make(chan bool),
		log:      logrus.WithFields(logrus.Fields{"caller": ModuleName}),
		cfg:      cfg,
		tags:     tags,
		jobChan:  make(chan DownloadRequest, MaxDownloadsInFlight),
	}

	go dl.updateIDs()
	go dl.MessagePipe()

	for id := 0; id < cfg.Download.NumWorkers; id++ {
		go dl.Worker(id)
	}

	return dl, nil
}

func (dl *Download) GetRecvChan() chan mq.Message {
	return dl.recvChan
}

func (dl *Download) GetSendChan() chan mq.Message {
	return dl.sendChan
}

func (dl *Download) MessagePipe() {
	dl.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-dl.recvChan:
			{
				msgType := msg.GetMsgType()
				switch msgType {
				case mq.MsgDownload:
					{
						downloadMsg := msg.GetContent().(*mq.DownloadMessage)
						dl.fetch(downloadMsg.Site, downloadMsg.ID, downloadMsg.Submitter)
					}
				}
			}
		case <-dl.quit:
			{
				return
			}
		}
	}
}

func (dl *Download) Worker(id int) {
	wrkLog := dl.log.WithFields(logrus.Fields{
		"worker": id,
	})

	wrkLog.Debug("Starting worker")
	for job := range dl.jobChan {
		var (
			stdout, stderr bytes.Buffer
		)

		wrkLog.WithFields(logrus.Fields{
			"site":      SiteToString[job.Site],
			"id":        job.ID,
			"submitter": job.Submitter,
		}).Debug("Accepted download request")

		// Check if file is already downloaded
		alreadyDownloaded := false
		for _, id := range dl.idList {
			if id == job.ID {
				alreadyDownloaded = true
				break
			}
		}
		if alreadyDownloaded {
			wrkLog.WithFields(logrus.Fields{
				"site":      SiteToString[job.Site],
				"id":        job.ID,
				"submitter": job.Submitter,
			}).Info("Track already downloaded")
			continue
		}

		// Check if song is too long for stream
		if !job.isAllowedLength() {
			wrkLog.Debug("Sending SongTooLong")
			stlMsg := mq.NewSongTooLongMessage(job.ID, job.Submitter)
			msg := mq.NewMessage(ModuleName, mq.MsgSongTooLong, stlMsg)
			dl.sendChan <- *msg

			continue
		}

		// Download track
		wrkLog.WithFields(logrus.Fields{
			"site":      SiteToString[job.Site],
			"id":        job.ID,
			"submitter": job.Submitter,
		}).Info("Downloading track")

		daMsg := mq.NewDownloadAcceptedMessage(job.ID, job.Submitter)
		msg := mq.NewMessage(ModuleName, mq.MsgDownloadAccepted, daMsg)
		dl.sendChan <- *msg

		outputFile := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", dl.cfg.Paths.TmpDir)
		url := fmt.Sprintf("%s%s", dl.cfg.Download.Url, job.ID)

		cmd := exec.Command(
			dl.cfg.Paths.Youtubedl,
			"-x",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--add-metadata",
			"--metadata-from-title", "%(artist)s - %(title)s",
			"-o", outputFile, url)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		cmdlog := wrkLog.WithFields(logrus.Fields{
			"cmd": cmd,
		})

		if err := cmd.Start(); err != nil {
			cmdlog.Warnf("cmd.Start: %v", err)
			continue
		}

		if err := cmd.Wait(); err != nil {
			cmdlog.Warnf("cmd.Wait: %v", err)
			continue
		}

		globPattern := fmt.Sprintf("%s/*-%s.mp3", dl.cfg.Paths.TmpDir, job.ID)
		results, err := filepath.Glob(globPattern)
		if err != nil {
			wrkLog.Warnf("filepath.Glob: %v", err)
			continue
		}

		fname := results[0]

		// Set the submitter of the track
		dl.tags.SetSubmitter(fname, job.Submitter)

		// Set the rating to default
		dl.tags.SetRating(fname, id3tags.RatingDefault)

		// Copy file into music directory
		name := path.Base(fname)
		dest := fmt.Sprintf("%s/%s", dl.cfg.Paths.Music, name)

		err = dl.copyFile(fname, dest)
		if err != nil {
			wrkLog.Warnf("dl.copyFile: %v", err)
		}

		if err := os.Remove(fname); err != nil {
			wrkLog.Warnf("os.Remove: %v", err)
		}

		// Notify MPD
		addMsg := mq.NewAddToDBMessage(dest)
		msg = mq.NewMessage(ModuleName, mq.MsgAddToDB, addMsg)
		dl.sendChan <- *msg

		// Notify user
		complMsg := mq.NewDownloadCompletedMessage(name[:len(name)-16], job.Submitter)
		msg = mq.NewMessage(ModuleName, mq.MsgDownloadCompleted, complMsg)
		dl.sendChan <- *msg

		wrkLog.WithFields(logrus.Fields{
			"fname":     dest,
			"submitter": job.Submitter,
		}).Debug("Track downloaded")
	}
}

func (dl *Download) copyFile(src, dst string) error {
	dl.log.WithFields(logrus.Fields{
		"src": src,
		"dst": dst,
	}).Debug("Copying file")

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %v", err)
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("ioutil.WriteFile: %v", err)
	}

	return nil
}

func (dr *DownloadRequest) isAllowedLength() bool {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", dr.ID)

	log := logrus.WithFields(logrus.Fields{
		"caller":    "DownloadRequest",
		"site":      SiteToString[dr.Site],
		"id":        dr.ID,
		"submitter": dr.Submitter,
	})

	resp, err := http.Get(url)
	if err != nil {
		log.Warnf("http.Get: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("ioutil.ReadAll: %v", err)
		return false
	}

	results := reSongLength.FindAllStringSubmatch(string(body), -1)

	duration := -1
	if len(results) == 0 {
		log.Warnf("No results found")
		return false
	}

	duration, err = strconv.Atoi(results[0][1])
	if err != nil {
		log.Warnf("strconv.Atoi: %v", err)
		return false
	}

	if duration > MaxSongLength {
		log.WithFields(logrus.Fields{
			"duration": duration,
		}).Warnf("Song too long")
		return false
	}

	log.WithFields(logrus.Fields{
		"duration": duration,
	}).Debug("Song length accepted")
	return true
}

func (dl *Download) updateIDs() {
	fs, err := os.Stat(dl.cfg.Paths.Music)
	if err != nil {
		dl.log.Warnf("os.Stat: %v", err)
		return
	}

	if !fs.IsDir() {
		dl.log.Warnf("os.Stat %s: Not a directory", dl.cfg.Paths.Music)
		return
	}

	files, err := ioutil.ReadDir(dl.cfg.Paths.Music)
	if err != nil {
		dl.log.Warnf("ioutil.ReadDir: %v", err)
		return
	}

	for _, fs = range files {
		fname := fs.Name()
		result := reIDs[SiteYoutube].FindAllStringSubmatch(fname, -1)
		if len(result) != 1 {
			dl.log.WithFields(logrus.Fields{
				"fname": fname,
			}).Warnf("Unable to parse %s ID", SiteToString[SiteYoutube])
			continue
		}
		id := result[0][1]
		dl.idList = append(dl.idList, id)
	}

	dl.log.WithFields(logrus.Fields{
		"numtracks": len(dl.idList),
	}).Debug("Updated ID list")

	dl.log.WithFields(logrus.Fields{
		"numtracks": len(dl.idList),
	}).Debug("Sending UpdateIDs")
	updateIDsMsg := mq.NewUpdateIDsMessage(dl.idList)
	msg := mq.NewMessage(ModuleName, mq.MsgUpdateIDs, updateIDsMsg)
	dl.sendChan <- *msg
}

func (dl *Download) fetch(site int, id, submitter string) {
	switch site {
	case SiteYoutube:
		{
			dl.log.WithFields(logrus.Fields{
				"site":      SiteToString[site],
				"id":        id,
				"submitter": submitter,
			}).Debug("Submitting download job")

			dl.jobChan <- DownloadRequest{
				Site:      site,
				ID:        id,
				Submitter: submitter,
			}
		}
	default:
		{
			dl.log.WithFields(logrus.Fields{
				"site":      site,
				"id":        id,
				"submitter": submitter,
			}).Warn("Unknown download site")
		}
	}
}
