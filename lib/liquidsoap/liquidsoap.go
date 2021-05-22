package liquidsoap

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/r3boot/test/lib/config"
)

type PlaylistEntry struct {
	Filename string
}

type LSClient struct {
	cfg          *config.LSConfig
	datastoreCfg *config.DatastoreConfig
	conn         net.Conn
	mux          sync.Mutex
}

func NewLSClient(cfg *config.LSConfig) (*LSClient, error) {
	client := &LSClient{
		cfg: cfg,
		mux: sync.Mutex{},
	}

	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("Connect: %v", err)
	}

	return client, nil
}

func (ls *LSClient) Connect() error {
	var err error
	ls.conn, err = net.Dial("tcp", ls.cfg.Address)
	if err != nil {
		ls.conn = nil
		return fmt.Errorf("net.Dial: %v", err)
	}

	return nil
}

func (ls *LSClient) RunCommand(cmd string) (string, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	if ls.conn == nil {
		return "", fmt.Errorf("not connected to liquidsoap")
	}
	fmt.Fprintf(ls.conn, cmd+"\n")

	reader := bufio.NewReader(ls.conn)
	buffer := bytes.Buffer{}
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if string(line) != "END" {
			buffer.Write(line)
			buffer.Write([]byte("\n"))
		}

		if isPrefix {
			break
		}
		if string(line) == "END" {
			if buffer.Len() > 0 {
				break
			}
		}
	}

	return buffer.String(), nil
}

func (ls *LSClient) Next() error {
	// 2600nl(dot)mp3.skip
	cmd := ls.cfg.OutputName + ".skip"
	_, err := ls.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("LSClient.RunCommand: %v", err)
	}
	return nil
}

func (ls *LSClient) NowPlaying() (*PlaylistEntry, error) {
	// 2600nl(dot)mp3.metadata
	cmd := ls.cfg.OutputName + ".metadata"

	output, err := ls.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("LSClient.RunCommand: %v", err)
	}

	// The output of metadata is a stack; Do parsing in the ugly way
	nowPlayingFilename := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "filename=") {
			nowPlayingFilename = strings.Split(strings.Split(line, "=")[1], "\"")[1]
		}
	}

	entry := &PlaylistEntry{
		Filename: nowPlayingFilename,
	}

	return entry, nil
}

func (ls *LSClient) GetPlaylist() ([]*PlaylistEntry, error) {
	// ls /music
	return nil, nil
}

func (ls *LSClient) GetQueue() ([]*PlaylistEntry, error) {
	/*
	 * for RID in request.queue; do
	 *   request.metadata RID | grep filename
	 * done
	 */
	cmd := "request.queue"

	output, err := ls.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("LSClient.RunCommand: %v", err)
	}

	playlistEntries := []*PlaylistEntry{}

	rids := strings.Split(output, " ")
	for _, rid := range rids {
		cmd := "request.metadata " + rid
		result, err := ls.RunCommand(cmd)
		if err != nil {
			return nil, fmt.Errorf("LSClient.RunCommand: %v", err)
		}

		for _, line := range strings.Split(result, "\n") {
			if strings.HasPrefix(line, "filename=") {
				filename := strings.Split(strings.Split(line, "=")[1], "\"")[1]
				entry := PlaylistEntry{
					Filename: filename,
				}
				playlistEntries = append(playlistEntries, &entry)
			}
		}
	}

	return playlistEntries, nil
}

func (ls *LSClient) Enqueue(track *PlaylistEntry) error {
	// request.push track
	cmd := "request.push " + track.Filename
	_, err := ls.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("LSClient.RunCommand: %v", err)
	}

	return nil
}
