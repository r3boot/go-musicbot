package dbclient

import (
	"fmt"
	"strings"
	"time"
)

type Track struct {
	Yid       string        `sql:"type:varchar(11),pk"`
	Filename  string        `sql:"filename"`
	Rating    int64         `sql:"rating"`
	Submitter string        `sql:"submitter"`
	Duration  float64       `sql:"duration"`
	AddedOn   time.Time     `sql:"added_on"`
	AlbumArt  string        `sql:"-"`
	Elapsed   time.Duration `sql:"-"`
	tableName struct{}      `pg:",discard_unknown_columns"`
}

func (obj *Track) String() string {
	return fmt.Sprintf("Track<%s %s %d %s %.02f %s>", obj.Yid, obj.Filename, obj.Rating, obj.Submitter, obj.Duration, obj.AddedOn)
}

func (obj *Track) Save() error {
	err := client.db.Update(obj)
	if err != nil {
		return fmt.Errorf("db.Update: %v", err)
	}

	return nil
}

func (obj *Track) Add() error {
	err := client.db.Insert(obj)
	if err != nil {
		return fmt.Errorf("db.Insert: %v", err)
	}

	return nil
}

func (obj *Track) Remove() error {
	err := client.db.Delete(obj)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return fmt.Errorf("Track.Remove db.Delete: %v", err)
	}

	return nil
}

func (db *DbClient) Search(q string) ([]Track, error) {
	tracks := make([]Track, 0)

	track, err := db.GetTrackByYid(q)
	if err == nil {
		tracks = append(tracks, *track)
		return tracks, nil
	}

	_, err = db.db.Query(&tracks, "WITH query_results AS (SELECT *, filename <-> ? AS dist FROM tracks) SELECT * FROM query_results WHERE dist < 0.9", q)
	if err != nil {
		return nil, fmt.Errorf("Query: %v", err)
	}

	return tracks, nil
}

func (db *DbClient) GetTrackByYid(yid string) (*Track, error) {
	track := &Track{}

	err := db.db.Model(track).Where("yid = ?", yid).Select()
	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	return track, nil
}
