package youtubeclient

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"bytes"

	"github.com/r3boot/go-musicbot/lib/id3tags"
)

func (yt *YoutubeClient) DownloadWorker(id int, downloadMetas <-chan DownloadMeta) {
	for meta := range downloadMetas {
		log.Infof("Downloading new YID: %s (submitted by %s)", meta.Yid, meta.Nickname)
		_, err := yt.DownloadYID(meta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			continue
		}
		// yt.SetMetadataForNewSong(meta, fileName)
	}
}

func (yt *YoutubeClient) SetMetadataForNewSong(meta DownloadMeta, fileName string) error {
	if err := yt.addYID(meta.Yid); err != nil {
		return fmt.Errorf("YoutubeClient.SetMetadataForNewSong: %v", err)
	}

	curRating, err := yt.id3.GetRating(fileName)
	if err != nil {
		return fmt.Errorf("YoutubeClient.SetMetadataForNewSong: %v", err)
	}

	if curRating != id3tags.RATING_UNKNOWN {
		return fmt.Errorf("YoutubeClient.SetMetadataForNewSong: Rating already set for %s", curRating)
	}

	_, err = yt.id3.SetRating(fileName, id3tags.RATING_DEFAULT)
	if err != nil {
		return fmt.Errorf("YoutubeClient.SetMetadataForNewSong: %v", err)
	}

	err = yt.id3.SetSubmitter(fileName, meta.Nickname)
	if err != nil {
		return fmt.Errorf("YoutubeClient.SetMetadataForNewSong: %v", err)
	}

	return nil
}

func (yt *YoutubeClient) HasYID(yid string) bool {
	yt.seenFileMutex.RLock()
	defer yt.seenFileMutex.RUnlock()

	fd, err := os.Open(yt.config.Youtube.SeenFile)
	if err != nil {
		log.Warningf("YoutubeClient.HasYID os.Open: %v", err)
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
		return fmt.Errorf("YoutubeClient.addYID os.OpenFile: %v", err)
	}
	defer fd.Close()

	if _, err = fd.WriteString(yid); err != nil {
		return fmt.Errorf("YoutubeClient.addYID fd.WriteString: %v", err)
	}

	return nil
}

func (yt *YoutubeClient) IsAllowedLength(yid string) (bool, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", yid)

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("YoutubeClient.IsAllowedLength http.Get: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("YoutubeClient.IsAllowedLength ioutil.ReadAll: %v", err)
	}

	results := reSongLength.FindAllStringSubmatch(string(body), -1)

	duration := -1
	if len(results) == 0 {
		return false, fmt.Errorf("YoutubeClient.IsAllowedLength: No results found for %s", yid)
	}

	duration, err = strconv.Atoi(results[0][1])
	if err != nil {
		return false, fmt.Errorf("YoutubeClient.IsAllowedLength strconv.Atoi: %v", err)
	}

	log.Infof("Duration: %d", duration)

	return duration <= MaxSongLength, nil
}

func (yt *YoutubeClient) DownloadYID(meta DownloadMeta) (string, error) {
	var stdout, stderr bytes.Buffer

	if yt.config.Youtube.NumWorkers <= 1 {
		yt.downloadMutex.Lock()
		defer yt.downloadMutex.Unlock()
	}

	if yt.HasYID(meta.Yid) {
		return "", fmt.Errorf("YoutubeClient.DownloadYID: YID %s has already been downloaded", meta.Yid)
	}

	isAllowedLength, err := yt.IsAllowedLength(meta.Yid)
	if err != nil {
		return "", fmt.Errorf("YoutubeClient.DownloadYID: %v", err)
	}

	if !isAllowedLength {
		return "", fmt.Errorf("YoutubeClient.DownloadYID: Song too lengthy")
	}

	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", yt.MusicDir)
	url := fmt.Sprintf("%s%s", yt.config.Youtube.BaseUrl, meta.Yid)
	cmd := exec.Command(
		yt.config.Youtube.Downloader,
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--add-metadata",
		"--metadata-from-title", "%(artist)s - %(title)s",
		"-o", output, url)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debugf("YoutubeClient.DownloadYID: Running command: %v", cmd)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("YoutubeClient.DownloadYID cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		msg := fmt.Sprintf("youtube-dl returned non-zero exit code\nStdout: %s\nStderr: %s\n", stdout.String(), stderr.String())
		return "", fmt.Errorf(msg)
	}

	globPattern := fmt.Sprintf("%s/*-%s.mp3", yt.MusicDir, meta.Yid)

	results, err := filepath.Glob(globPattern)
	if err != nil {
		return "", fmt.Errorf("YoutubeClient.DownloadYID filepath.Glob: %v", err)
	}

	if results == nil {
		return "", fmt.Errorf("YoutubeClient.DownloadYID: filepath.Glob did not return any results")
	}

	fileName := results[0]

	yt.SetMetadataForNewSong(meta, fileName)

	if yt.mpdClient != nil {
		yt.mpdMutex.Lock()
		defer yt.mpdMutex.Unlock()
		if err := yt.mpdClient.UpdateDB(filepath.Base(fileName)); err != nil {
			log.Warningf("YoutubeClient.DownloadYID: %v", err)
		}
	}

	return results[0], nil
}
