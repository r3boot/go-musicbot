package discogs

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/r3boot/go-musicbot/lib/log"
)

const (
	searchUrl   = "https://api.discogs.com/database/search?q="
	notFoundUri = "notfound.png"
)

type dgPaginationUrls struct {
	Last string `json:"last"`
	Next string `json:"next"`
}

type dgPagination struct {
	PerPage int              `json:"per_page"`
	Pages   int              `json:"pages"`
	Page    int              `json:"page"`
	Urls    dgPaginationUrls `json:"urls"`
}

type dgResultCommunity struct {
	Want int `json:"want"`
	Have int `json:"have"`
}

type dgResult struct {
	Style       []string          `json:"style"`
	Thumb       string            `json:"thumb"`
	Format      []string          `json:"format"`
	Country     string            `json:"country"`
	Barcode     []string          `json:"barcode"`
	Uri         string            `json:"uri"`
	Community   dgResultCommunity `json:"community"`
	Label       []string          `json:"label"`
	CatNo       string            `json:"catno"`
	Year        string            `json:"year"`
	Genre       []string          `json:"genre"`
	Title       string            `json:"title"`
	ResourceUrl string            `json:"resource_url"`
	Type        string            `json:"type"`
	Id          int               `json:"id"`
}

type dgSearchResult struct {
	Pagination dgPagination `json:"pagination"`
	Items      int          `json:"items"`
	Results    []dgResult   `json:"results"`
}

type Discogs struct {
	cfg *config.Discogs
}

func sha1sum(input string) (string, error) {
	hasher := sha1.New()
	nwritten, err := hasher.Write([]byte(input))
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "sha1sum",
			"call":     "hasher.Write",
		}, err.Error())
		return "", fmt.Errorf("failed to write hash")
	}
	if nwritten != len(input) {
		log.Warningf(log.Fields{
			"package":     "discogs",
			"function":    "sha1sum",
			"num_written": nwritten,
			"input_size":  len(input),
		}, "corrupt write")
		return "", fmt.Errorf("corrupt write")
	}

	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return hash, nil
}

func sanitizeQuery(q string) string {
	result := ""

	insideBrackets := false
	for _, chr := range q {
		if chr == '[' || chr == '(' || chr == '{' {
			insideBrackets = true
			continue
		}
		if chr == ']' || chr == ')' || chr == '}' {
			insideBrackets = false
			continue
		}
		if insideBrackets {
			continue
		}
		result += string(chr)
	}

	return result
}

func NewDiscogs(cfg *config.Discogs) *Discogs {
	discogs := &Discogs{
		cfg: cfg,
	}

	return discogs
}

func (a *Discogs) HasImage(hash string) (bool, error) {
	files, err := ioutil.ReadDir(a.cfg.CacheDir)
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "Discogs",
			"function":  "HasImage",
			"call":      "ioutil.ReadDir",
			"cache_dir": a.cfg.CacheDir,
		}, err.Error())
		return false, fmt.Errorf("failed to read directory")
	}

	imgFile := fmt.Sprintf("%s.jpg", hash)
	for _, fs := range files {
		if fs.IsDir() {
			continue
		}
		if !strings.HasSuffix(fs.Name(), ".jpg") {
			continue
		}

		if fs.Name() == imgFile {
			return true, nil
		}
	}

	return false, nil
}

func (a *Discogs) DownloadImage(imgUrl, hash string) error {
	localFile := fmt.Sprintf("%s/%s.jpg", a.cfg.CacheDir, hash)

	fd, err := os.Create(localFile)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "Download",
			"call":     "os.Create",
			"filename": localFile,
		}, err.Error())
		return fmt.Errorf("failed to create file")
	}
	defer fd.Close()

	log.Debugf(log.Fields{
		"package":  "discogs",
		"function": "Download",
		"url":      imgUrl,
		"hash":     hash,
	}, "downloading album art")

	client := &http.Client{}

	request, err := http.NewRequest("GET", imgUrl, nil)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "Download",
			"call":     "http.NewRequest",
			"url":      imgUrl,
		}, err.Error())
		return fmt.Errorf("failed to create http request")
	}

	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")

	response, err := client.Do(request)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "Download",
			"call":     "client.Do",
		}, err.Error())
		return fmt.Errorf("failed to submit http request")
	}
	defer response.Body.Close()

	ncopied, err := io.Copy(fd, response.Body)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "Download",
			"call":     "io.Copy",
		}, err.Error())
		return fmt.Errorf("failed to save body")
	}

	log.Debugf(log.Fields{
		"package":  "discogs",
		"function": "Download",
		"filename": localFile,
		"size":     ncopied,
	}, "saved album art")
	return nil
}

func (a *Discogs) StoreLocally(imgUrl, hash string) error {
	hasImage, err := a.HasImage(hash)
	if err != nil {
		return fmt.Errorf("HasImage: %v", err)
	}

	if hasImage {
		return nil
	}

	return a.DownloadImage(imgUrl, hash)
}

func (a *Discogs) GetAlbumArt(query string) (string, error) {
	saneQuery := sanitizeQuery(query)

	queryUrl := searchUrl + url.QueryEscape(saneQuery)
	hash, err := sha1sum(query)
	if err != nil {
		return "", fmt.Errorf("sha1sum: %v", err)
	}

	imgUrl := fmt.Sprintf("%s.jpg", hash)

	hasImage, err := a.HasImage(hash)
	if err != nil {
		return "", fmt.Errorf("HasImage: %v", err)
	}
	if hasImage {
		log.Debugf(log.Fields{
			"package":  "discogs",
			"function": "GetAlbumArt",
			"call":     "HasImage",
			"img_url":  imgUrl,
		}, "found cached album art")
		return imgUrl, nil
	}

	log.Debugf(log.Fields{
		"package":  "discogs",
		"function": "GetAlbumArt",
		"query":    query,
	}, "searching for album art")

	client := &http.Client{}

	request, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "GetAlbumArt",
			"call":     "http.NewRequest",
			"url":      "queryUrl",
		}, err.Error())
		return "", fmt.Errorf("failed to create http request")
	}

	authValue := fmt.Sprintf("Discogs token=%s", a.cfg.Token)
	request.Header.Add("Authorization", authValue)

	response, err := client.Do(request)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "GetAlbumArt",
			"call":     "client.Do",
		}, err.Error())
		return "", fmt.Errorf("failed to submit http request")
	}
	defer response.Body.Close()

	rawData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "GetAlbumArt",
			"call":     "ioutil.ReadAll",
		}, err.Error())
		return "", fmt.Errorf("failed to read body")
	}

	responseData := dgSearchResult{}
	err = json.Unmarshal(rawData, &responseData)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "discogs",
			"function": "GetAlbumArt",
			"call":     "json.Unmarshal",
		}, err.Error())
		return "", fmt.Errorf("failed to unmarshal json")
	}

	if len(responseData.Results) > 0 {
		err = a.StoreLocally(responseData.Results[0].Thumb, hash)
		if err != nil {
			return "", fmt.Errorf("StoreLocally: %v", err)
		}
		return imgUrl, nil
	}

	log.Warningf(log.Fields{
		"package":  "discogs",
		"function": "GetAlbumArt",
		"query":    query,
	}, "no results found")
	return notFoundUri, nil
}
