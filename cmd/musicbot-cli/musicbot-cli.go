package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"bytes"
	"regexp"
	"time"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/ytclient"
)

const (
	D_DEBUG         = false
	D_GET_RATING    = true
	D_SET_RATING    = -1
	D_BASEDIR       = "/music/2600nl"
	D_CFGFILE       = "/etc/musicbot.yaml"
	D_PLAYLISTDIR   = "/var/lib/mpd/2600nl/playlists"
	D_PL_ELITE      = false
	D_PL_FAVOURITES = false
	D_FETCH_MISSING = ""
	D_RECURSE       = true
	D_USE_TIMESTAMP = false
)

var (
	debug              = flag.Bool("D", D_DEBUG, "Enable debug mode")
	cfgFile            = flag.String("f", D_CFGFILE, "Path to configuration file")
	getRating          = flag.Bool("lr", D_GET_RATING, "Show current rating")
	setRating          = flag.Int("sr", D_SET_RATING, "Set the rating")
	baseDir            = flag.String("d", D_BASEDIR, "Music directory")
	playlistFavourites = flag.Bool("pf", D_PL_FAVOURITES, "Generate playlist for 6 or higher rating")
	playlistElite      = flag.Bool("pe", D_PL_ELITE, "Generate playlist for tracks with rating 9 or higher")
	playlistDir        = flag.String("pd", D_PLAYLISTDIR, "Directory to store playlists in")
	fetchMissing       = flag.String("fetch-missing", D_FETCH_MISSING, "Fetch all YIDs not in list")
	RE_YID             = regexp.MustCompile(".*([a-zA-Z0-9_-]{11}).mp3")
	Id3Tags            *id3tags.ID3Tags
	YoutubeClient      *youtubeclient.YoutubeClient
	Config             *config.MusicBotConfig
	Logger             *logger.Logger
)

func SetRating(fname string, newRating int) {
	fullPath, err := filepath.Abs(fname)
	if err != nil {
		Logger.Fatalf("SetRating filepath.Abs: %v", err)
	}

	_, err = Id3Tags.SetRating(fullPath, newRating)
	if err != nil {
		Logger.Fatalf("SetRating: %v", err)
	}
}

func ShowRating(fname string) {
	fullPath, err := filepath.Abs(fname)
	if err != nil {
		Logger.Fatalf("ShowRating filepath.Abs: %v", err)
	}

	rating, err := Id3Tags.GetRating(fullPath)
	if err != nil {
		Logger.Fatalf("ShowRating: %v", err)
	}

	fmt.Printf("%d %s\n", rating, fullPath)
}

func ShowRatingsForDir(dirname string) {
	entries, err := ioutil.ReadDir(dirname)
	if err != nil {
		Logger.Warningf("GetRatingsForDir ioutil.ReadDir: %v", err)
		return
	}

	for _, entry := range entries {
		target := dirname + "/" + entry.Name()
		if entry.IsDir() {
			ShowRatingsForDir(target)
		} else {
			ShowRating(target)
		}
	}
}

// Playlist based on a 6 or higher rating
func GeneratePlayList(name string, minRating int) {
	ratings, err := Id3Tags.GetAllRatings()
	if err != nil {
		Logger.Fatalf("GeneratePlaylist: %v", err)
	}

	favourites := []string{}

	for path, rating := range ratings {
		if rating < minRating {
			continue
		}

		fullPath := fmt.Sprintf("%s/%s", *baseDir, path)
		favourites = append(favourites, fullPath)
	}

	fname := fmt.Sprintf("%s/%s.m3u", *playlistDir, name)

	fd, err := os.Create(fname)
	if err != nil {
		Logger.Fatalf("GenerateFavouritesPlayList os.Open: %v", err)
	}
	defer fd.Close()

	for _, entry := range favourites {
		data := fmt.Sprintf("%s\n", entry)
		fd.WriteString(data)
	}

	Logger.Infof("Wrote %s\n", fname)
}

// Fetch all yids not found in list
func FetchMissing(inputFile string) {
	wantedYid := []string{}

	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		Logger.Fatalf("FetchMissing ioutil.ReadFile: %v", err)
	}

	for _, rawYid := range bytes.Split(data, []byte("\n")) {
		yid := string(rawYid)
		if yid == "" {
			continue
		}

		wantedYid = append(wantedYid, yid)
	}

	entries, err := ioutil.ReadDir(*baseDir)
	if err != nil {
		Logger.Fatalf("FetchMissing ioutil.ReadDir: %v", err)
	}

	haveYid := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		result := RE_YID.FindAllStringSubmatch(entry.Name(), -1)
		if len(result) > 0 {
			yid := result[0][1]
			haveYid = append(haveYid, yid)
		}
	}

	toFetchYid := []string{}
	for _, wanted := range wantedYid {
		hasYid := false
		for _, have := range haveYid {
			if have == wanted {
				hasYid = true
			}
		}
		if !hasYid {
			toFetchYid = append(toFetchYid, wanted)
		}
	}

	for _, yid := range toFetchYid {
		YoutubeClient.DownloadChan <- yid
		Logger.Infof("Added %v to download queue", yid)
	}

	for {
		queueSize := len(YoutubeClient.DownloadChan)
		if queueSize == 0 {
			break
		}
		Logger.Infof("Queue has %d items", queueSize)
		time.Sleep(5 * time.Second)
	}
}

func init() {
	flag.Parse()

	Logger = logger.NewLogger(D_USE_TIMESTAMP, *debug)
	Config, err := config.LoadConfig(Logger, *cfgFile)
	if err != nil {
		Logger.Fatalf("config.LoadConfig: %v", err)
	}

	Id3Tags = id3tags.NewID3Tags(Logger, *baseDir)
	YoutubeClient = youtubeclient.NewYoutubeClient(Logger, Config, nil, Id3Tags, *baseDir)
}

func main() {
	target := ""
	if len(flag.Args()) > 0 {
		target = flag.Args()[0]
	}

	if *setRating != -1 {
		if target == "" {
			Logger.Fatalf("Need a target")
		}
		SetRating(target, *setRating)
	} else if *playlistFavourites {
		GeneratePlayList("favourites", 7)
	} else if *playlistElite {
		GeneratePlayList("elite", 9)
	} else if *fetchMissing != D_FETCH_MISSING {
		FetchMissing(*fetchMissing)
	} else if *getRating {
		if target == "" {
			Logger.Fatalf("Need a target")
		}
		fs, err := os.Stat(target)
		if err != nil {
			Logger.Fatalf("os.Stat: %v", err)
		}

		if fs.IsDir() {
			ShowRatingsForDir(target)
		} else {
			ShowRating(target)
		}
	}
}
