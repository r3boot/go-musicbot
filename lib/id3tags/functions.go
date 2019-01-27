package id3tags

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (t *TagList) Has(fname string) bool {
	tagListMutex.RLock()
	defer tagListMutex.RUnlock()

	if strings.HasPrefix(fname, "/") {
		fname = path.Base(fname)
	}

	for key, _ := range *t {
		if key == fname {
			return true
		}
	}
	return false
}

func (i *ID3Tags) SetRating(fname string, rating int) (int, error) {
	if rating <= RATING_ZERO || rating >= RATING_MAX {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.SetRating: rating must be between %d and %d", RATING_ZERO, RATING_MAX)
	}

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.SetRating: %v", err)
	}

	rating_s := strconv.Itoa(rating)

	params := []string{"-T", rating_s, fullPath}
	_, err = i.RunId3v2(params)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.SetRating: %v", err)
	}

	return rating, nil
}

func (i *ID3Tags) GetRating(fname string) (int, error) {
	var rating int

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.GetRating: %v", err)
	}

	rating = RATING_UNKNOWN
	result, _ := i.ReadFrame(fullPath, id3Track)

	if result != "" {
		rating, err = strconv.Atoi(result)
		if err != nil {
			return RATING_UNKNOWN, fmt.Errorf("ID3Tags.GetRating strconv.Atoi:%v", err)
		}
	} else {
		rating = RATING_UNKNOWN
	}

	return rating, nil
}

func (i *ID3Tags) DecreaseRating(name string) (int, error) {
	fullPath, err := i.expandFullPath(name)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: %v", err)
	}

	rating, err := i.GetRating(fullPath)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: %v", err)
	}

	switch rating {
	case RATING_UNKNOWN:
		return RATING_UNKNOWN, nil
	case RATING_ZERO:
		return RATING_ZERO, nil
	default:
		{
			rating -= 1
			rating, err := i.SetRating(fullPath, rating)
			if err != nil {
				return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating:%v", err)
			}
			return rating, nil
		}
	}

	return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: unknown error")
}

func (i *ID3Tags) IncreaseRating(name string) (int, error) {
	fullPath, err := i.expandFullPath(name)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
	}

	rating, err := i.GetRating(fullPath)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
	}

	switch rating {
	case RATING_UNKNOWN:
		return RATING_UNKNOWN, nil
	case RATING_MAX:
		return RATING_MAX, nil
	default:
		{
			rating += 1
			rating, err := i.SetRating(name, rating)
			if err != nil {
				return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
			}
			return rating, nil
		}
	}

	return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: unknown error")
}

func getYID(filename string) (string, error) {
	fnameLen := len(filename)

	if fnameLen < 15 {
		return "", fmt.Errorf("getYID: filename is too short")
	}

	return filename[fnameLen-15:fnameLen-4], nil
}

func isTrack(line string) (string, bool) {
	result := reTrack.FindAllStringSubmatch(line, -1)

	if len(result) == 0 {
		return "", false
	}

	return result[0][1], true
}

func isArtist(line string) (string, bool) {
	result := reArtist.FindAllStringSubmatch(line, -1)

	if len(result) == 0 {
		return "", false
	}

	return result[0][1], true
}

func isTitle(line string) (string, bool) {
	result := reTitle.FindAllStringSubmatch(line, -1)

	if len(result) == 0 {
		return "", false
	}

	return result[0][1], true
}

func (i *ID3Tags) UpdateTags() error {
	destPath := fmt.Sprintf("%s/*.mp3", i.BaseDir)

	for {
		time.Sleep(5 * time.Second)

		allEntries, err := filepath.Glob(destPath)
		if err != nil {
			return fmt.Errorf("ID3Tags.GetTags filepath.Glob: %v", err)
		}
		log.Debugf("ID3Tags.UpdateTags: Found %d files", len(allEntries))

		filteredEntries := []string{}
		for _, entry := range allEntries {
			if i.tagList.Has(entry) {
				continue
			}
			filteredEntries = append(filteredEntries, entry)
		}

		params := []string{"-l"}
		params = append(params, filteredEntries...)
		output, err := i.RunId3v2(params)
		if err != nil {
			return fmt.Errorf("ID3Tags.GetTags: %v", err)
		}

		curTrack := ""
		numIndexed := 0
		tagListMutex.Lock()
		for _, line := range strings.Split(output, "\n") {
			if line == "" {
				continue
			}
			if fullTrack, ok := isTrack(line); ok {
				curTrack = path.Base(fullTrack)
				yid, err := getYID(curTrack)
				if err != nil {
					log.Warningf("ID3Tags.GetTags getYID: Failed: %v", err)
					continue
				}
				newTrack := &TrackTags{
					Filename: curTrack,
					YID: yid,
				}
				i.tagList[curTrack] = newTrack
			}

			if artist, ok := isArtist(line); ok {
				if curTrack == "" {
					log.Warningf("ID3Tags.GetTags isArtist: Track is empty")
					continue
				}

				i.tagList[curTrack].Artist = artist
			}

			if title, ok := isTitle(line); ok {
				if curTrack == "" {
					log.Warningf("ID3Tags.GetTags isArtist: Track is empty")
					tagListMutex.Unlock()
					continue
				}
				i.tagList[curTrack].Title = title
			}
		}
		tagListMutex.Unlock()

		batch := i.searchIdx.NewBatch()
		for _, track := range i.tagList {
			batch.Index(track.YID, track)
			numIndexed += 1
		}
		i.searchIdx.Batch(batch)

		log.Infof("ID3Tags.GetTags: Added %d tracks to search index", numIndexed)
	}

	return nil
}

func (i *ID3Tags) GetTags() (TagList, error) {
	return i.tagList, nil
}

func (i *ID3Tags) RemoveFile(fname string) error {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return fmt.Errorf("ID3Tags.RemoveFile: %v", err)
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.RemoveFile os.Stat: %v", err)
	}

	err = os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.RemoveFile os.Remove: %v", err)
	}

	return nil
}
func (i *ID3Tags) GetAllFiles() ([]string, error) {
	files, err := ioutil.ReadDir(i.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("ID3Tags.GetAllFiles ioutil.ReadDir: %v", err)
	}

	tmpList := make([]string, len(files))
	totItems := 0
	for _, fs := range files {
		if fs.IsDir() {
			continue
		} else if fs.Name() == "" {
			continue
		} else if !strings.HasSuffix(fs.Name(), ".mp3") {
			continue
		}

		tmpList = append(tmpList, fs.Name())
		totItems += 1
	}

	response := make([]string, totItems)
	response = tmpList

	sort.Strings(response)

	return response, nil
}

func (i *ID3Tags) GetAllRatings() (map[string]int, error) {
	ratings := map[string]int{}

	files, err := ioutil.ReadDir(i.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("ID3Tags.GetAllRatings ioutil.ReadDir: %v", err)
	}

	for _, fs := range files {
		if fs.IsDir() {
			continue
		}

		if fs.Name() == "" {
			continue
		}

		fullPath, err := i.expandFullPath(fs.Name())
		if err != nil {
			return nil, fmt.Errorf("ID3Tags.GetAllRatings: %v", err)
		}

		rating, err := i.GetRating(fullPath)
		if err != nil {
			return nil, fmt.Errorf("ID3Tags.GetAllRatings: %v", err)
		}

		ratings[fs.Name()] = rating
	}

	return ratings, nil
}

func (i *ID3Tags) GetAuthor(fname string) (string, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.GetAuthor: %v", err)
	}

	result, err := i.ReadFrame(fullPath, id3Artist)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.GetAuthor: %v", err)
	}

	return result, nil
}

func (i *ID3Tags) GetTitle(fname string) (string, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.GetTitle: %v", err)
	}

	result, _ := i.ReadFrame(fullPath, id3Title)
	return result, nil
}

func (i *ID3Tags) GetSubmitter(fname string) (string, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.SetSubmitter: %v", err)
	}

	submitter, _ := i.ReadFrame(fullPath, id3Comment)

	return submitter, nil
}

func (i *ID3Tags) HasSubmitter(fname string) (bool, error) {
	result, err := i.GetSubmitter(fname)
	if err != nil {
		return false, fmt.Errorf("ID3Tags.HasSubmitter: %v", err)
	}

	return len(result) > 0, nil
}

func (i *ID3Tags) SetSubmitter(fname, nickname string) error {
	hasSubmitter, err := i.HasSubmitter(fname)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetSubmitter: %v", err)
	}

	if hasSubmitter {
		return fmt.Errorf("ID3Tags.SetSubmitter: track already has a submitter")
	}

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetSubmitter: %v", err)
	}

	nickname = strings.TrimSpace(nickname)

	_, err = i.RunId3v2([]string{"-c", nickname, fullPath})
	if err != nil {
		return fmt.Errorf("ID3Tags.SetSubmitter: %v", err)
	}

	return nil
}

func (i *ID3Tags) SetMetadata(fname string) error {
	var artist, title string

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	title, err = i.GetTitle(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	if title != "" {
		return fmt.Errorf("ID3Tags.SetMetadata: Metadata already set")
	}

	rating, err := i.GetRating(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	if rating != RATING_UNKNOWN {
		return fmt.Errorf("ID3Tags.SetMetadata: Rating already set already set")
	}

	name := path.Base(fullPath)
	name = name[:len(name)-16]
	rating = 5

	if strings.Count(name, "-") >= 1 {
		tokens := strings.Split(name, "-")
		artist = strings.TrimSpace(tokens[0])
		title = strings.TrimSpace(tokens[1])
	} else {
		title = name
	}

	params := []string{}
	rating_s := strconv.Itoa(rating)

	if len(artist) > 0 {
		params = []string{
			"-a", artist,
			"-t", title,
			"-T", rating_s,
			fullPath,
		}
	} else {
		params = []string{
			"-t", title,
			"-T", rating_s,
			fullPath,
		}
	}

	_, err = i.RunId3v2(params)
	if err != nil {
		fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	return nil
}

func (i *ID3Tags) OpenSearchIndex(idxFile string) error {
	mapping := bleve.NewIndexMapping()

	_, err := os.Stat(idxFile)

	if err != nil {
		i.searchIdx, err = bleve.New(idxFile, mapping)
		if err != nil {
			return fmt.Errorf("ID3Tags.New: %v", err)
		}
	} else {
		i.searchIdx, err = bleve.Open(idxFile)
		if err != nil {
			return fmt.Errorf("ID3Tags.New: %v", err)
		}
	}

	return nil
}