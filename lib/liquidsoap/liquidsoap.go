package liquidsoap

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/log"
	"io"
	"net"
	"path"
	"strings"
	"sync"
	"time"
)

type PlaylistEntry struct {
	Filename string
	Elapsed  time.Duration
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
		return nil, fmt.Errorf("failed to connect")
	}

	return client, nil
}

func (ls *LSClient) Connect() error {
	var err error
	ls.conn, err = net.Dial("tcp", ls.cfg.Address)
	if err != nil {
		ls.conn = nil
		log.Warningf(log.Fields{
			"package":  "liquidsoap",
			"function": "Connect",
			"call":     "net.Dial",
			"address":  ls.cfg.Address,
		}, err.Error())
		return fmt.Errorf("failed to connect")
	}

	return nil
}

func (ls *LSClient) RunCommand(cmd string) (string, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	if ls.conn == nil {
		log.Warningf(log.Fields{
			"package":  "liquidsoap",
			"function": "RunCommand",
			"command":  cmd,
		}, "not connected to liquidsoap")
		return "", fmt.Errorf("failed to run command")
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
		log.Warningf(log.Fields{
			"package":  "liquidsoap",
			"function": "Next",
			"call":     "ls.RunCommand",
		}, err.Error())
		return fmt.Errorf("output.skip failed")
	}
	return nil
}

func (ls *LSClient) NowPlaying() (*PlaylistEntry, error) {
	// 2600nl(dot)mp3.metadata
	cmd := ls.cfg.OutputName + ".metadata"

	output, err := ls.RunCommand(cmd)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "liquidsoap",
			"function": "NowPlaying",
			"call":     "ls.RunCommand",
		}, err.Error())
		return nil, fmt.Errorf("output.metadata failed")
	}

	// The output of metadata is a stack; Do parsing in the ugly way
	nowPlayingFilename := ""
	onAir := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "filename=") {
			nowPlayingFilename = strings.Split(strings.Split(line, "=")[1], "\"")[1]
		}
		if strings.HasPrefix(line, "on_air=") {
			onAir = strings.Split(strings.Split(line, "=")[1], "\"")[1]
		}
	}

	tStart, err := time.Parse("2006/01/02 15:04:05", onAir)
	if err != nil {
		log.Warningf(log.Fields{
			"package":   "liquidsoap",
			"function":  "NowPlaying",
			"call":      "time.Parse",
			"timestamp": onAir,
		}, err.Error())
	}

	entry := &PlaylistEntry{
		Filename: nowPlayingFilename,
		Elapsed:  time.Since(tStart),
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
		log.Warningf(log.Fields{
			"package":  "liquidsoap",
			"function": "NowPlaying",
			"call":     "ls.RunCommand",
		}, err.Error())
		return nil, fmt.Errorf("request.queue failed")
	}

	playlistEntries := []*PlaylistEntry{}

	rids := strings.Split(output, " ")
	for _, rid := range rids {
		cmd := "request.metadata " + rid
		result, err := ls.RunCommand(cmd)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "liquidsoap",
				"function": "NowPlaying",
				"call":     "ls.RunCommand",
			}, err.Error())
			return nil, fmt.Errorf("request.metadata failed")
		}

		for _, line := range strings.Split(result, "\n") {
			if strings.HasPrefix(line, "filename=") {
				filename := path.Base(strings.Split(strings.Split(line, "=")[1], "\"")[1])
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
		log.Warningf(log.Fields{
			"package":  "liquidsoap",
			"function": "NowPlaying",
			"call":     "ls.RunCommand",
		}, err.Error())
		return fmt.Errorf("request.push failed")
	}

	return nil
}
