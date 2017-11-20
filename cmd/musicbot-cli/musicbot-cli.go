package main

import (
	"flag"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/logger"
	"github.com/r3boot/go-musicbot/lib/mp3lib"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	D_DEBUG         = false
	D_GET_RATING    = true
	D_SET_RATING    = -1
	D_BASEDIR       = "/music"
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
func GenerateFavouritesPlayList() {
	ratings := MP3Library.GetAllRatings()

	favourites := []string{}

	for path, rating := range ratings {
		if rating < 6 {
			continue
		}
		favourites = append(favourites, path)
	}

	fmt.Printf("%v\n", favourites)
}

func init() {
	flag.Parse()

	Logger = logger.NewLogger(D_USE_TIMESTAMP, *debug)
	MP3Library = mp3lib.NewMP3Library(Logger, *baseDir)

	// No arguments passed
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	target := flag.Args()[0]
	if *setRating != -1 {
		SetRating(target, *setRating)
	} else if *playlistFavourites {
		GenerateFavouritesPlayList()
	} else if *getRating {
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
