package dbclient

import (
	"fmt"
	"strings"
	"time"
)

type Track struct {
	Id        int
	Filename  string
	Yid       string
	Rating    int64
	Submitter string
	Duration  float64
	Elapsed   float64 `pg:"-"`
	Priority  int     `pg:"-"`
	AddedOn   time.Time
	tableName struct{} `pg:",discard_unknown_columns"`
}

func (obj *Track) String() string {
	return fmt.Sprintf("Track<%d %s %d %s %.02f %s>", obj.Id, obj.Filename, obj.Rating, obj.Submitter, obj.Duration, obj.AddedOn)
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

	_, err := db.db.Query(&tracks, "SELECT *, filename <-> ? AS dist FROM tracks ORDER BY dist ASC LIMIT 10", q)
	if err != nil {
		return nil, fmt.Errorf("Query: %v", err)
	}

	return tracks, nil
}

func (db *DbClient) GetTrackById(id int) (*Track, error) {
	track := &Track{}

	err := db.db.Model(track).Where("id = ?", id).Select()
	if err != nil {
		return nil, fmt.Errorf("Query: %v", err)
	}

	return track, nil
}

func (db *DbClient) GetTrackByYid(yid string) (*Track, error) {
	track := &Track{}

	err := db.db.Model(track).Where("yid = ?", yid).Select()
	if err != nil {
		return nil, fmt.Errorf("Query: %v", err)
	}

	return track, nil
}
