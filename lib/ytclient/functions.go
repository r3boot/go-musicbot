package youtubeclient

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bytes"
	"strings"

	"github.com/r3boot/go-musicbot/lib/mp3lib"
)

func (yt *YoutubeClient) DownloadSerializer() {
	for {
		newYid := <-yt.DownloadChan
		fmt.Printf("Downloading new YID: %s\n", newYid)
		fileName, err := yt.DownloadYID(newYid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			continue
		}
		yt.SetMetadataForNewSong(newYid, fileName)
	}
}

func (yt *YoutubeClient) PlaylistSerializer() {
	for {
		newPlaylistUrl := <-yt.PlaylistChan
		fmt.Printf("Downloading playlist: %s\n", newPlaylistUrl)
		yt.DownloadPlaylist(newPlaylistUrl)
	}
}

func (yt *YoutubeClient) SetMetadataForNewSong(yid, fileName string) error {
	if err := yt.addYID(yid); err != nil {
		return fmt.Errorf("SetMetadataForNewSong: Failed to add yid to seen file: %v", err)
	}

	curRating := yt.mp3Library.GetRating(fileName)
	if curRating != mp3lib.RATING_UNKNOWN {
		return fmt.Errorf("SetMetadataForNewSong: Rating already set for %s", curRating)
	}
	yt.mp3Library.SetRating(fileName, mp3lib.RATING_DEFAULT)

	return nil
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

func (yt *YoutubeClient) DownloadYID(yid string) (string, error) {
	var stdout, stderr bytes.Buffer

	yt.downloadMutex.Lock()
	defer yt.downloadMutex.Unlock()

	if yt.hasYID(yid) {
		return "", fmt.Errorf("YID %s has already been downloaded\n", yid)
	}

	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", yt.musicDir)
	url := fmt.Sprintf("%s%s", yt.config.Youtube.BaseUrl, yid)
	cmd := exec.Command(yt.config.Youtube.Downloader, "-x", "--audio-format", "mp3", "-o", output, url)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("Running command: %v\n", cmd)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Failed to run %s: %v\n", yt.config.Youtube.Downloader, err)
	}

	if err := cmd.Wait(); err != nil {
		msg := fmt.Sprintf("youtube-dl returned non-zero exit code\nStdout: %s\nStderr: %s\n", stdout.String(), stderr.String())
		return "", fmt.Errorf(msg)
	}

	globPattern := fmt.Sprintf("%s/*-%s.mp3", yt.musicDir, yid)
	fmt.Printf("globPattern: %v\n", globPattern)
	results, err := filepath.Glob(globPattern)
	if err != nil {
		return "", fmt.Errorf("YoutubeClient.DownloadYID: %v\n", err)
	}

	if results == nil {
		return "", fmt.Errorf("YoutubeClient.DownloadYID: filepath.Glob did not return any results\n")
	}

	fileName := results[0]

	yt.SetMetadataForNewSong(yid, fileName)

	if err := yt.mpdClient.UpdateDB(filepath.Base(fileName)); err != nil {
		fmt.Printf("Failed to update mpd database: %v\n", err)
	}

	return results[0], nil
}

func (yt *YoutubeClient) DownloadPlaylist(url string) error {
	var stdout, stderr bytes.Buffer

	yt.downloadMutex.Lock()
	defer yt.downloadMutex.Unlock()

	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", yt.musicDir)
	cmd := exec.Command(yt.config.Youtube.Downloader, "-x", "-i", "--audio-format", "mp3", "-o", output, url)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("Running command: %v\n", cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to run %s: %v\n", yt.config.Youtube.Downloader, err)
	}

	if err := cmd.Wait(); err != nil {
		msg := fmt.Sprintf("youtube-dl returned non-zero exit code\nStdout: %s\nStderr: %s\n", stdout.String(), stderr.String())
		return fmt.Errorf(msg)
	}

	for _, line := range strings.Split(stdout.String(), "\n") {

		if !strings.Contains(line, "Destination:") {
			continue
		}

		result := RE_DESTINATION.FindAllStringSubmatch(line, -1)

		if len(result) != 1 {
			continue
		}

		yid := result[0][2]
		fileName := result[0][1] + "-" + yid + ".mp3"

		yt.SetMetadataForNewSong(yid, fileName)

		if err := yt.mpdClient.UpdateDB(filepath.Base(fileName)); err != nil {
			return fmt.Errorf("Failed to update mpd database: %v\n", err)
		}
	}

	return nil
}
