package mpdclient

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fhs/gompd/mpd"
)

func (m *MPDClient) Connect() error {
	var err error

	if m.config.MPD.Password != "" {
		m.conn, err = mpd.DialAuthenticated("tcp", m.address, m.config.MPD.Password)
		if err != nil {
			return fmt.Errorf("connect failed: %v", err)
		}
	} else {
		m.conn, err = mpd.Dial("tcp", m.address)
		if err != nil {
			return fmt.Errorf("connect failed: %v", err)
		}
	}

	return nil
}

func (m *MPDClient) KeepAlive() {
	var err error

	for {
		if m.conn == nil { // Socket is closed, connect to mpd again
			if err = m.Connect(); err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
		}

		if err = m.conn.Ping(); err != nil { // Ping command failed, reconnect to mpd
			m.Close()
			if err = m.Connect(); err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
		}

		time.Sleep(time.Second * 3)
	}
}

func (m *MPDClient) UpdateDB(fname string) error {
	_, err := m.conn.Update(fname)
	time.Sleep(1 * time.Second)
	return err
}

func (m *MPDClient) Close() error {
	var err error

	if err = m.conn.Close(); err != nil {
		return fmt.Errorf("MPD.Close failed: %v\n", err)
	}

	return nil
}

func (m *MPDClient) NowPlaying() string {
	attrs, err := m.conn.CurrentSong()
	if err != nil {
		return fmt.Sprintf("Error: Failed to fetch current song info: %v", err)
	}
	return attrs["file"]
}

func (m *MPDClient) Duration() string {
	attrs, err := m.conn.CurrentSong()
	if err != nil {
		return fmt.Sprintf("Error: Failed to fetch current song info: %v", err)
	}

	rawDuration := strings.Split(attrs["duration"], ".")[0]
	rawDuration += "s"
	duration, err := time.ParseDuration(rawDuration)
	if err != nil {
		return fmt.Sprintf("Error: Failed to parse duration: %v", err)
	}
	return duration.String()
}

func (m *MPDClient) Next() string {
	m.conn.Next()
	return m.NowPlaying()
}

func (m *MPDClient) Play() string {
	m.Shuffle()
	m.conn.Play(1)
	return m.NowPlaying()
}

func (m *MPDClient) PlayPos(pos int) string {
	m.conn.Play(pos)
	return m.NowPlaying()
}

func (m *MPDClient) Shuffle() {
	m.conn.Shuffle(-1, -1)
}

func (m *MPDClient) Search(q string) (int, error) {
	result, err := m.conn.Search("filename", q)
	if err != nil {
		return -1, fmt.Errorf("MPDClient.Search: %v", err)
	}

	if len(result) == 0 {
		return -1, fmt.Errorf("MPDClient.Search: no songs found")
	}

	curPlaylist, err := m.conn.PlaylistInfo(-1, -1)
	if err != nil {
		return -1, fmt.Errorf("MPDClient.Search: failed to retrieve playlist")
	}

	for _, song := range curPlaylist {
		if song["file"] == string(result[0]["file"]) {
			pos, err := strconv.Atoi(song["Pos"])
			if err != nil {
				return -1, fmt.Errorf("MPDClient.Search: failed to convert pos to int")
			}
			return pos, nil
		}
	}

	return -1, fmt.Errorf("MPDClient.Search: failed to search mpd")
}
