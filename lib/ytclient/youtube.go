package youtubeclient

import (
	"bufio"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/mpdclient"
	"os"
	"os/exec"
	"sync"
	"time"
)

func NewYoutubeClient(config *config.MusicBotConfig, mpdclient *mpdclient.MPDClient, musicDir string) *YoutubeClient {
	yt := &YoutubeClient{
		seenFileMutex: sync.RWMutex{},
		downloadMutex: sync.RWMutex{},
		config:        config,
		mpdClient:     mpdclient,
		musicDir:      musicDir,
	}

	return yt
}

func (yt *YoutubeClient) hasYID(yid string) bool {
	yt.seenFileMutex.RLock()
	defer yt.seenFileMutex.RUnlock()

	fd, err := os.Open(yt.config.Youtube.SeenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		return false
	}

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if scanner.Text() == yid {
			return true
		}
	}

	return false
}

func (yt *YoutubeClient) addYID(yid string) error {
	yt.seenFileMutex.Lock()
	defer yt.seenFileMutex.Unlock()

	fd, err := os.OpenFile(yt.config.Youtube.SeenFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("addYID failed: %v", err)
	}
	defer fd.Close()

	if _, err = fd.WriteString(yid); err != nil {
		return fmt.Errorf("addYID failed to add yid to file: %v", err)
	}

	return nil
}

func (yt *YoutubeClient) DownloadYID(yid string) {
	yt.downloadMutex.Lock()
	defer yt.downloadMutex.Unlock()

	if yt.hasYID(yid) {
		fmt.Printf("YID %s has already been downloaded\n", yid)
	}
	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", yt.musicDir)
	url := fmt.Sprintf("%s%s", yt.config.Youtube.BaseUrl, yid)
	cmd := exec.Command(yt.config.Youtube.Downloader, "-x", "--audio-format", "mp3", "-o", output, url)
	fmt.Printf("Running command: %v\n", cmd)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to run %s: %v\n", yt.config.Youtube.Downloader, err)
	}
	cmd.Wait()

	if err := yt.addYID(yid); err != nil {
		fmt.Printf("Failed to add yid to seen file: %v\n", err)
	}

	/*
		if err := yt.mpdClient.UpdateDB(); err != nil {
			fmt.Printf("Failed to update mpd database: %v\n", err)
		}
		time.Sleep(1 * time.Second)
	*/
}
