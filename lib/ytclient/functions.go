package youtubeclient

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bytes"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
)

func (yt *YoutubeClient) DownloadSerializer() {
	for {
		newYid := <-yt.DownloadChan
		fmt.Printf("Downloading new YID: %s\n", newYid)
		fileName := yt.DownloadYID(newYid)
		yt.mp3Library.SetRating(fileName, mp3lib.RATING_DEFAULT)
		fmt.Printf("Set rating for %s to %d\n", fileName, mp3lib.RATING_DEFAULT)
	}
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

func (yt *YoutubeClient) DownloadYID(yid string) string {
	var stdout, stderr bytes.Buffer

	yt.downloadMutex.Lock()
	defer yt.downloadMutex.Unlock()

	if yt.hasYID(yid) {
		fmt.Printf("YID %s has already been downloaded\n", yid)
		return ""
	}

	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", yt.musicDir)
	url := fmt.Sprintf("%s%s", yt.config.Youtube.BaseUrl, yid)
	cmd := exec.Command(yt.config.Youtube.Downloader, "-x", "--audio-format", "mp3", "-o", output, url)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("Running command: %v\n", cmd)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to run %s: %v\n", yt.config.Youtube.Downloader, err)
		return ""
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "youtube-dl returned non-zero exit code\n")
		fmt.Fprintf(os.Stderr, "stdout: %s\n", stdout.String())
		fmt.Fprintf(os.Stderr, "stderr: %s\n", stderr.String())
		return ""
	}

	if err := yt.addYID(yid); err != nil {
		fmt.Printf("Failed to add yid to seen file: %v\n", err)
		return ""
	}

	/*
		if err := yt.mpdClient.UpdateDB(); err != nil {
			fmt.Printf("Failed to update mpd database: %v\n", err)
		}
	*/

	globPattern := fmt.Sprintf("%s/*-%s.mp3", yt.musicDir, yid)
	fmt.Printf("globPattern: %v\n", globPattern)
	results, err := filepath.Glob(globPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "YoutubeClient.DownloadYID: %v\n", err)
		return ""
	}

	if results == nil {
		fmt.Fprintf(os.Stderr, "YoutubeClient.DownloadYID: filepath.Glob did not return any results\n")
		return ""
	}

	return results[0]
}
