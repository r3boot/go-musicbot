package mp3lib

import (
	"fmt"
	"os"
	"strconv"

	id3 "github.com/mikkyang/id3-go"
	"io/ioutil"
	"sort"
)

func (i *MP3Library) SetRating(fname string, rating int) int {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	fd, err := id3.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MP3Library.GetRating %v\n", err)
		return RATING_UNKNOWN
	}
	defer fd.Close()

	newRating := strconv.Itoa(rating)
	fd.SetYear(newRating)

	fmt.Printf("Set rating for %s to %d\n", fname, RATING_DEFAULT)

	return rating
}

func (i *MP3Library) GetRating(fname string) int {
	var err error

	if _, err := os.Stat(fname); err != nil {
		fmt.Fprintf(os.Stderr, "MP3Library.GetRating %v\n", err)
		return RATING_UNKNOWN
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()

	fd, err := id3.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MP3Library.GetRating %v\n", err)
		return RATING_UNKNOWN
	}
	defer fd.Close()

	curRating_s := fd.Year()

	if curRating_s == "" {
		return RATING_UNKNOWN
	}

	curRating, err := strconv.Atoi(curRating_s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MP3Library.GetRating %v\n", err)
		return RATING_UNKNOWN
	}

	return curRating
}

func (i *MP3Library) DecreaseRating(name string) int {
	curRating := i.GetRating(name)

	switch curRating {
	case RATING_UNKNOWN:
		return RATING_UNKNOWN
	case RATING_ZERO:
		return RATING_ZERO
	default:
		{
			curRating -= 1

			rating := i.SetRating(name, curRating)

			return rating
		}
	}

	return RATING_UNKNOWN
}

func (i *MP3Library) IncreaseRating(name string) int {
	curRating := i.GetRating(name)

	switch curRating {
	case RATING_UNKNOWN:
		return RATING_UNKNOWN
	case RATING_MAX:
		return RATING_MAX
	default:
		{
			curRating += 1

			rating := i.SetRating(name, curRating)

			return rating
		}

	}

	return RATING_UNKNOWN
}

func (i *MP3Library) RemoveFile(name string) bool {
	var err error

	fname := i.BaseDir + "/" + name

	_, err = os.Stat(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MP3Library.RemoveFile %v\n", err)
		return false
	}

	err = os.Remove(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MP3Library.RemoveFile: %v\n", err)
		return false
	}

	return true
}

func (i *MP3Library) GetAllFiles() []string {

	files, err := ioutil.ReadDir(i.BaseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	tmpList := make([]string, len(files))
	totItems := 0
	for _, fs := range files {
		if fs.Name() == "" {
			continue
		}
		tmpList = append(tmpList, fs.Name())
		totItems += 1
	}

	response := make([]string, totItems)
	response = tmpList

	sort.Strings(response)

	return response
}
