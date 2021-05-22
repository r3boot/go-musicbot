package ytclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/r3boot/test/lib/config"
)

const (
	downloadQueueMaxLength = 100
)

var (
	// \"approxDurationMs\":\"464046\"
	reAllowedSongLength = regexp.MustCompile("approxDurationMs..:..([0-9]{4,10})..")
)

type DownloadJob struct {
	Yid       string
	Submitter string
}

type YoutubeClient struct {
	cfg       *config.YoutubeConfig
	Datastore string
}

func NewYoutubeClient(cfg *config.Config) (*YoutubeClient, error) {
	client := &YoutubeClient{
		cfg:       cfg.Youtube,
		Datastore: cfg.Datastore.Directory,
	}

	if client.cfg.Binary == "" {
		binary, err := client.FindBinary()
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		client.cfg.Binary = binary
	}

	return client, nil
}

func (yt *YoutubeClient) FindBinary() (string, error) {
	bindirs := []string{"/sbin", "/usr/sbin", "/usr/local/sbin", "/bin", "/usr/bin", "/usr/local/bin"}

	for _, dirName := range bindirs {
		fullPath := dirName + "/youtube-dl"
		fs, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if fs.IsDir() {
			continue
		}
		return fullPath, nil
	}

	return "", fmt.Errorf("FindBinary: youtube-dl not found")
}

func (yt *YoutubeClient) IsAllowedLength(yid string) error {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", yid)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("isAllowedLength: http.Get: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("isAllowedLength: ioutil.ReadAll: %v", err)
	}

	results := reAllowedSongLength.FindAllStringSubmatch(string(body), -1)

	duration := -1
	if len(results) == 0 {
		return fmt.Errorf("isAllowedLength: YID not found")
	}

	duration, err = strconv.Atoi(results[0][1])
	if err != nil {
		return fmt.Errorf("isAllowedLength: strconv.Atoi: %v", err)
	}

	if ((duration / 1000) / 60) > yt.cfg.MaxAllowedLength {
		return fmt.Errorf("isAllowedLength: song too long")
	}

	return nil
}

func (yt *YoutubeClient) copyFile(src, dst string) error {
	log := logrus.WithFields(logrus.Fields{
		"module":   "YoutubeClient",
		"function": "copyFile",
		"src":      src,
		"dst":      dst,
	})
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %v", err)
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("ioutil.WriteFile: %v", err)
	}

	log.Printf("File copied")

	return nil
}

func (yt *YoutubeClient) Download(job *DownloadJob) (string, error) {
	var (
		stdout, stderr bytes.Buffer
	)

	log := logrus.WithFields(logrus.Fields{
		"module":    "YoutubeClient",
		"function":  "Download",
		"yid":       job.Yid,
		"submitter": job.Submitter,
	})

	outputFile := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", yt.cfg.TmpDir)
	url := fmt.Sprintf("%s%s", yt.cfg.BaseUrl, job.Yid)

	cmd := exec.Command(
		yt.cfg.Binary,
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", outputFile, url)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Running %s", strings.Join(cmd.Args, " "))

	if err := cmd.Start(); err != nil {
		log.Printf("cmd.Start: %v", err)
		return "", fmt.Errorf("Failed to start youtube-dl")
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("cmd.Wait: %v", err)
		return "", fmt.Errorf("Youtube-dl did not complete")
	}

	globPattern := fmt.Sprintf("%s/*-%s.mp3", yt.cfg.TmpDir, job.Yid)
	results, err := filepath.Glob(globPattern)
	if err != nil {
		log.Printf("Failed to find downloaded file")
		return "", fmt.Errorf("Youtube-dl failed to download file")
	}
	fname := results[0]

	name := path.Base(fname)
	dest := fmt.Sprintf("%s/%s", yt.Datastore, name)

	err = yt.copyFile(fname, dest)
	if err != nil {
		log.Printf("copyFile: %v", err)
		return "", fmt.Errorf("Failed to copy file")
	}

	if err := os.Remove(fname); err != nil {
		log.Printf("Remove: %v", err)
		return "", fmt.Errorf("Failed to remove tmpfile")
	}

	log.Printf("Track downloaded succesfully")

	return name, nil
}
