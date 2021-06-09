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

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/log"
)

const (
	ytUrl = "https://www.youtube.com/watch?v=%s"
)

var (
	reAllowedSongLength = regexp.MustCompile("approxDurationMs\":\"([0-9]{4,10})\"")
	reSongTitle         = regexp.MustCompile("<title>(.*) - YouTube</title>")
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
			log.Fatalf(log.Fields{
				"package":  "ytclient",
				"function": "NewYoutubeClient",
				"call":     "client.FindBinary",
			}, err.Error())
			return nil, fmt.Errorf("%v", err)
		}
		client.cfg.Binary = binary
	}

	log.Debugf(log.Fields{
		"package":  "ytclient",
		"function": "NewYoutubeClient",
		"binary":   client.cfg.Binary,
	}, "found binary")

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

	return "", fmt.Errorf("youtube-dl not found")
}

func (yt *YoutubeClient) SongTitle(yid string) (string, error) {
	url := fmt.Sprintf(ytUrl, yid)

	resp, err := http.Get(url)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "SongTitle",
			"call":     "http.Get",
			"yid":      yid,
		}, err.Error())
		return "", fmt.Errorf("failed to get url")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "SongTitle",
			"call":     "ioutil.ReadAll",
			"yid":      yid,
		}, err.Error())
		return "", fmt.Errorf("failed to read url")
	}

	results := reSongTitle.FindAllStringSubmatch(string(body), -1)
	if len(results) == 0 {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "SongTitle",
			"yid":      yid,
		}, "no results found")
		return "", fmt.Errorf("no results found")
	}

	title := results[0][1]

	return title, nil
}

func (yt *YoutubeClient) IsAllowedLength(yid string) error {
	url := fmt.Sprintf(ytUrl, yid)

	resp, err := http.Get(url)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "IsAllowedLength",
			"call":     "http.Get",
			"yid":      yid,
		}, err.Error())
		return fmt.Errorf("failed to get url")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "IsAllowedLength",
			"call":     "ioutil.ReadAll",
			"yid":      yid,
		}, err.Error())
		return fmt.Errorf("failed to read url")
	}

	results := reAllowedSongLength.FindAllStringSubmatch(string(body), -1)

	duration := -1
	if len(results) == 0 {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "IsAllowedLength",
			"yid":      yid,
		}, "no results found")
		return fmt.Errorf("no results found")
	}

	duration, err = strconv.Atoi(results[0][1])
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "IsAllowedLength",
			"call":     "strconv.Atoi",
			"duration": duration,
			"yid":      yid,
		}, "failed to convert duration")
		return fmt.Errorf("failed to convert duration")
	}

	if ((duration / 1000) / 60) > yt.cfg.MaxAllowedLength {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "IsAllowedLength",
			"duration": duration,
			"yid":      yid,
		}, "track too long")
		return fmt.Errorf("track too long")
	}

	return nil
}

func (yt *YoutubeClient) copyFile(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "copyFile",
			"call":     "ioutil.ReadFile",
			"src":      src,
		}, err.Error())
		return fmt.Errorf("failed to read file")
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "copyFile",
			"call":     "ioutil.WriteFile",
			"dst":      dst,
		}, err.Error())
		return fmt.Errorf("failed to write file")
	}

	log.Debugf(log.Fields{
		"package":  "ytclient",
		"function": "copyFile",
		"call":     "ioutil.WriteFile",
		"src":      src,
		"dst":      dst,
	}, "file copied")

	return nil
}

func (yt *YoutubeClient) Download(job *DownloadJob) (string, error) {
	var (
		stdout, stderr bytes.Buffer
	)

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

	command := strings.Join(cmd.Args, " ")

	log.Debugf(log.Fields{
		"package":  "ytclient",
		"function": "Download",
		"command":  command,
	}, "running command")

	if err := cmd.Start(); err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "Download",
			"call":     "cmd.Start",
			"command":  command,
		}, err.Error())
		return "", fmt.Errorf("failed to start youtube-dl")
	}

	if err := cmd.Wait(); err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "Download",
			"call":     "cmd.Wait",
			"command":  command,
		}, err.Error())
		return "", fmt.Errorf("youtube-dl did not complete")
	}

	globPattern := fmt.Sprintf("%s/*-%s.mp3", yt.cfg.TmpDir, job.Yid)
	results, err := filepath.Glob(globPattern)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "Download",
			"call":     "filepath.Glob",
			"pattern":  globPattern,
		}, err.Error())
		return "", fmt.Errorf("youtube-dl failed to download file")
	}
	fname := results[0]

	name := path.Base(fname)
	dest := fmt.Sprintf("%s/%s", yt.Datastore, name)

	err = yt.copyFile(fname, dest)
	if err != nil {
		return "", fmt.Errorf("failed to copy file")
	}

	if err := os.Remove(fname); err != nil {
		log.Warningf(log.Fields{
			"package":  "ytclient",
			"function": "Download",
			"call":     "os.Remove",
			"filename": fname,
		}, err.Error())
		return "", fmt.Errorf("failed to remove tmpfile")
	}

	log.Debugf(log.Fields{
		"package":   "ytclient",
		"function":  "Download",
		"yid":       job.Yid,
		"submitter": job.Submitter,
		"filename":  dest,
	}, "file downloaded succesfully")

	return name, nil
}
