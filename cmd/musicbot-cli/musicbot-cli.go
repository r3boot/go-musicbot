package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
)

const (
	D_DEBUG         = false
	D_GET_RATING    = true
	D_SET_RATING    = -1
	D_BASEDIR       = "/music/2600nl"
	D_PLAYLISTDIR   = "/var/lib/mpd/2600nl/playlists"
	D_PL_ELITE      = false
	D_PL_FAVOURITES = false
	D_RECURSE       = true
	D_USE_TIMESTAMP = false
)

var (
	debug              = flag.Bool("D", D_DEBUG, "Enable debug mode")
	getRating          = flag.Bool("lr", D_GET_RATING, "Show current rating")
	setRating          = flag.Int("sr", D_SET_RATING, "Set the rating")
	baseDir            = flag.String("d", D_BASEDIR, "Music directory")
	playlistFavourites = flag.Bool("pf", D_PL_FAVOURITES, "Generate playlist for 6 or higher rating")
	playlistElite      = flag.Bool("pe", D_PL_ELITE, "Generate playlist for tracks with rating 9 or higher")
	playlistDir        = flag.String("pd", D_PLAYLISTDIR, "Directory to store playlists in")
	MP3Library         *mp3lib.MP3Library
	Logger             *logger.Logger
)

func SetRating(fname string, newRating int) {
	fullPath, err := filepath.Abs(fname)
	if err != nil {
		Logger.Fatalf("SetRating filepath.Abs: %v", err)
	}

	MP3Library.SetRating(fullPath, newRating)
}

func ShowRating(fname string) {
	fullPath, err := filepath.Abs(fname)
	if err != nil {
		Logger.Fatalf("ShowRating filepath.Abs: %v", err)
	}

	rating := MP3Library.GetRating(fullPath)

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
	ratings := MP3Library.GetAllRatings()

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

func init() {
	flag.Parse()

	Logger = logger.NewLogger(D_USE_TIMESTAMP, *debug)
	MP3Library = mp3lib.NewMP3Library(Logger, *baseDir)
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
