package tags

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/tcolgate/mp3"

	"github.com/r3boot/test/lib/dbclient"

	"github.com/dhowden/tag"
)

var (
	reYid = regexp.MustCompile(".*([a-zA-Z0-9_-]{11}).mp3$")
)

func GetDuration(fname string) (float64, error) {
	duration := 0.0

	fd, err := os.Open(fname)
	if err != nil {
		return -1, fmt.Errorf("os.Open: %v", err)
	}

	decoder := mp3.NewDecoder(fd)
	var f mp3.Frame
	skipped := 0

	for {
		err := decoder.Decode(&f, &skipped)
		if err != nil {
			if err == io.EOF {
				break
			}
			return -1, fmt.Errorf("decoder.Decode: %v", err)
		}

		duration += f.Duration().Seconds()
	}

	return duration, nil
}

func ReadTagsFrom(pattern string) (result []dbclient.Track, err error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("filepath.Glob: %v\n", err)
	}

	fmt.Printf("Indexing files: ")

	ctr := 0
	for _, fname := range files {
		results := reYid.FindAllStringSubmatch(fname, -1)
		if len(results) == 0 {
			fmt.Printf("no yid found for %s\n", fname)
			continue
		}
		yid := results[0][1]

		fs, err := os.Stat(fname)
		if err != nil {
			fmt.Printf("os.Stat: %v\n", err)
			continue
		}
		fd, err := os.Open(fname)
		if err != nil {
			fmt.Printf("os.Open: %v\n", err)
			continue
		}
		m, err := tag.ReadFrom(fd)
		if err != nil {
			fmt.Printf("tag.ReadFrom: %v\n", err)
			continue
		}

		raw := m.Raw()
		comment := ""
		if raw["COMM"] != nil {
			comment = raw["COMM"].(*tag.Comm).Text
		}

		_, ok := raw["TLEN"]
		duration := 0.0
		if ok {
			rawDuration := raw["TLEN"].(string)
			duration, err = strconv.ParseFloat(rawDuration, 10)
			if err != nil {
				fmt.Printf("strconv.ParseFloat: %v\n", err)
				continue
			}
		} else {
			duration, err = GetDuration(fname)
			if err != nil {
				fmt.Printf("GetDuration: %v\n", err)
				continue
			}
		}

		track := dbclient.Track{
			Filename:  path.Base(fname),
			Yid:       yid,
			Submitter: comment,
			Duration:  duration,
			AddedOn:   fs.ModTime(),
		}

		fd.Close()

		result = append(result, track)

		ctr += 1
		if ctr > 10 {
			fmt.Printf(".")
			ctr = 0
		}
	}

	fmt.Printf("\nFound %d tracks\n", len(result))

	return result, nil
}
