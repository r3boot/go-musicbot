package tags

import (
	"fmt"
	"github.com/r3boot/go-musicbot/lib/log"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/tcolgate/mp3"

	"github.com/dhowden/tag"
	"github.com/r3boot/go-musicbot/lib/dbclient"
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
		log.Fatalf(log.Fields{
			"package":  "tags",
			"function": "ReadTagsFrom",
			"call":     "filepath.Glob",
			"pattern":  pattern,
		}, err.Error())
		return nil, fmt.Errorf("filepath.Glob: %v\n", err)
	}

	log.Debugf(log.Fields{
		"package":  "tags",
		"function": "ReadTagsFrom",
	}, "scanning files")

	tStart := time.Now()
	for _, fname := range files {
		results := reYid.FindAllStringSubmatch(fname, -1)
		if len(results) == 0 {
			log.Warningf(log.Fields{
				"package":  "tags",
				"function": "ReadTagsFrom",
				"call":     "reYid.FindAllStringSubmatch",
				"filename": fname,
			}, err.Error())
			continue
		}
		yid := results[0][1]

		fs, err := os.Stat(fname)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "tags",
				"function": "ReadTagsFrom",
				"call":     "os.Stat",
				"filename": fname,
			}, err.Error())
			continue
		}
		fd, err := os.Open(fname)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "tags",
				"function": "ReadTagsFrom",
				"call":     "os.Open",
				"filename": fname,
			}, err.Error())
			continue
		}
		m, err := tag.ReadFrom(fd)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "tags",
				"function": "ReadTagsFrom",
				"call":     "tag.ReadFrom",
				"filename": fname,
			}, err.Error())
			continue
		}

		raw := m.Raw()
		comment := ""
		if raw["COMM"] != nil {
			comment = raw["COMM"].(*tag.Comm).Text
		}

		rating, _ := m.Track()

		_, ok := raw["TLEN"]
		duration := 0.0
		if ok {
			rawDuration := raw["TLEN"].(string)
			duration, err = strconv.ParseFloat(rawDuration, 10)
			if err != nil {
				log.Warningf(log.Fields{
					"package":  "tags",
					"function": "ReadTagsFrom",
					"call":     "strconv.ParseFloat",
					"duration": rawDuration,
				}, err.Error())
				continue
			}
		} else {
			duration, err = GetDuration(fname)
			if err != nil {
				log.Warningf(log.Fields{
					"package":  "tags",
					"function": "ReadTagsFrom",
					"call":     "GetDuration",
					"filename": fname,
				}, err.Error())
				continue
			}
		}

		track := dbclient.Track{
			Filename:  path.Base(fname),
			Yid:       yid,
			Submitter: comment,
			Duration:  duration,
			Rating:    int64(rating),
			AddedOn:   fs.ModTime(),
		}

		fd.Close()

		result = append(result, track)
	}

	tEnd := time.Since(tStart)
	duration := fmt.Sprintf("%s", tEnd)

	log.Debugf(log.Fields{
		"package":   "tags",
		"function":  "ReadTagsFrom",
		"num_found": len(result),
		"duration":  duration,
	}, "scanning finished")

	return result, nil
}
