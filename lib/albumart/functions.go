package albumart

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func (a *AlbumArt) HasImage(hash string) (bool, error) {
	files, err := ioutil.ReadDir(a.cacheDir)
	if err != nil {
		return false, fmt.Errorf("AlbumArt.HasImage ioutil.ReadDir: %v", err)
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

func (a *AlbumArt) Download(imgUrl, hash string) error {
	localFile := fmt.Sprintf("%s/%s.jpg", a.cacheDir, hash)

	fd, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("AlbumArt.Download os.Create: %v", err)
	}
	defer fd.Close()

	log.Debugf("AlbumArt.Download: downloading from %s", imgUrl)
	client := &http.Client{}

	request, err := http.NewRequest("GET", imgUrl, nil)
	if err != nil {
		return fmt.Errorf("AlbumArt.Download http.NewRequest: %v", err)
	}

	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("AlbumArt.Download http.Get: %v", err)
	}
	defer response.Body.Close()

	ncopied, err := io.Copy(fd, response.Body)
	if err != nil {
		return fmt.Errorf("AlbumArt.Download io.Copy: %v", err)
	}

	log.Debugf("Wrote %d bytes to %s", ncopied, localFile)
	return nil
}

func (a *AlbumArt) StoreLocally(imgUrl, hash string) error {
	hasImage, err := a.HasImage(hash)
	if err != nil {
		return fmt.Errorf("AlbumArt.StoreLocally: %v", err)
	}

	if hasImage {
		return nil
	}

	return a.Download(imgUrl, hash)
}

func (a *AlbumArt) GetAlbumArt(query string) (string, error) {
	saneQuery := SantizeQuery(query)

	log.Infof("AlbumArt.GetAlbumArt: searching for %s", saneQuery)

	queryUrl := SEARCH_URL + url.QueryEscape(saneQuery)
	hash, err := sha1sum(query)
	if err != nil {
		return "", fmt.Errorf("AlbumArt.GetAlbumArt: %v", err)
	}
	imgUrl := fmt.Sprintf("/img/art/%s.jpg", hash)

	hasImage, err := a.HasImage(hash)
	if err != nil {
		return "", fmt.Errorf("AlbumArt.GetAlbumArt: %v", err)
	}
	if hasImage {
		return imgUrl, nil
	}

	client := &http.Client{}

	request, err := http.NewRequest("GET", queryUrl, nil)
	if err != nil {
		return "", fmt.Errorf("AlbumArt.GetAlbumArt http.NewRequest: %v", err)
	}

	request.Header.Add("Authorization", "Discogs token=GLXsrCZGDuAYdZrBibudYGBpXyAtBRtdmdzHkoaE")

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("AlbumArt.GetAlbumArt client.Do: %v", err)
	}
	defer response.Body.Close()

	rawData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("AlbumArt.GetAlbumArt ioutil.ReadAll: %v", err)
	}

	responseData := dgSearchResult{}
	err = json.Unmarshal(rawData, &responseData)
	if err != nil {
		return "", fmt.Errorf("AlbumArt.GetAlbumArt json.Unmarshal: %v", err)
	}

	if len(responseData.Results) > 0 {
		err = a.StoreLocally(responseData.Results[0].Thumb, hash)
		if err != nil {
			return "", fmt.Errorf("AlbumArt.GetAlbumArt: %v", err)
		}
		return imgUrl, nil
	}

	log.Warningf("AlbumArt.GetAlbumArt: nothing found for %s", query)
	return NOTFOUND_URI, nil
}
