package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/fhs/gompd/mpd"
	"github.com/thoj/go-ircevent"
	"gopkg.in/sevlyar/go-daemon.v0"
	"gopkg.in/yaml.v2"
)

const (
	D_CFGFILE string = "musicbot.yaml"

	CMD_DJPLUS  string = "dj+"
	CMD_NEXT    string = "next"
	CMD_PLAYING string = "np"
	CMD_RADIO   string = "radio"
)

type IrcConfig struct {
	Nickname  string `yaml:"nickname"`
	Server    string `yaml:"server"`
	Port      int    `yaml:"port"`
	Channel   string `yaml:"channel"`
	UseTLS    bool   `yaml:"tls"`
	Daemonize bool   `yaml:"daemonize"`
}

type BotConfig struct {
	CommandChar   string   `yaml:"command_character"`
	ValidCommands []string `yaml:"valid_commands"`
	StreamURL     string   `yaml:"stream_url"`
}

type YoutubeConfig struct {
	BaseDir    string `yaml:"music_basedir"`
	BaseUrl    string `yaml:"url"`
	Downloader string `yaml:"downloader"`
	SeenFile   string `yaml:"seen_file"`
}

type MpdConfig struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type MusicBotConfig struct {
	IRC     IrcConfig     `yaml:"irc"`
	Bot     BotConfig     `yaml:"bot"`
	Youtube YoutubeConfig `yaml:"youtube"`
	MPD     MpdConfig     `yaml:"mpd"`
}

type MPD struct {
	address string
	conn    *mpd.Client
}

var irccon *irc.Connection
var Config *MusicBotConfig
var MpdClient *MPD

var RE_CMD = regexp.MustCompile("^(\\![a-z\\+\\-]{2,6})")
var RE_DJHANDLER = regexp.MustCompile("(\\!dj\\+) ([a-zA-Z0-9_-]{11})")

var (
	cfgFile  = flag.String("f", D_CFGFILE, "Configuration file to use")
	musicDir string
)

func NewMPD() (*MPD, error) {
	client := &MPD{
		address: fmt.Sprintf("%s:%d", Config.MPD.Address, Config.MPD.Port),
	}

	return client, nil
}

func (m *MPD) Connect() error {
	var err error

	if Config.MPD.Password != "" {
		m.conn, err = mpd.DialAuthenticated("tcp", m.address, Config.MPD.Password)
		if err != nil {
			return fmt.Errorf("MPD.Connect failed: %v", err)
		}
	} else {
		m.conn, err = mpd.Dial("tcp", m.address)
		if err != nil {
			return fmt.Errorf("MPD.Connect failed: %v", err)
		}
	}

	return nil
}

func (m *MPD) Close() error {
	var err error

	if err = m.conn.Close(); err != nil {
		return fmt.Errorf("MPD.Close failed: %v\n", err)
	}

	return nil
}

func (m *MPD) NowPlaying() string {
	m.Connect()
	defer m.Close()

	attrs, err := m.conn.CurrentSong()
	if err != nil {
		return fmt.Sprintf("Error: Failed to fetch current song info: %v", err)
	}
	return attrs["file"]
}

func (m *MPD) Next() string {
	m.Connect()
	defer m.Close()
	return m.NowPlaying()
}

func LoadConfig(filename string) (config *MusicBotConfig, err error) {
	config = &MusicBotConfig{}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig failed: %v", err)
	}

	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("LoadConfig failed: %v", err)
	}

	return config, nil
}

func isValidCommand(cmd string) (string, bool) {
	cmdReString := fmt.Sprintf("^\\%s([a-z\\+\\-]{2,6})", Config.Bot.CommandChar)
	fmt.Printf("%v\n", cmdReString)
	reValidCmd := regexp.MustCompile(cmdReString)

	result := reValidCmd.FindAllStringSubmatch(cmd, -1)

	if len(result) == 0 {
		fmt.Printf("isValidCommand: %s does not match the valid command regexp\n", cmd)
		return "", false
	}

	wantedCommand := result[0][1]
	for _, validCommand := range Config.Bot.ValidCommands {
		if wantedCommand == validCommand {
			return wantedCommand, true
		}
	}

	fmt.Printf("isValidCommand: Unknown command: %s\n", cmd)
	return "", false
}

func hasYID(yid string) bool {
	fd, err := os.Open(Config.Youtube.SeenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		return false
	}

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if scanner.Text() == yid {
			return true
		}
	}

	return false
}

func stripChannel(channel string) string {
	result := ""
	for i := 0; i < len(channel); i++ {
		if channel[i] == '#' {
			continue
		}
		result += string(channel[i])
	}

	return result
}

func DownloadYID(yid string) {
	if hasYID(yid) {
		fmt.Printf("YID %s has already been downloaded\n", yid)
	}
	output := fmt.Sprintf("%s/%%(title)s-%%(id)s.%%(ext)s", musicDir)
	url := fmt.Sprintf("%s%s", Config.Youtube.BaseUrl, yid)
	cmd := exec.Command(Config.Youtube.Downloader, "-x", "--audio-format", "mp3", "-o", output, url)
	fmt.Printf("Running command: %v\n", cmd)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to run %s: %v\n", Config.Youtube.Downloader, err)
	}
	cmd.Wait()
}

func HandleYidDownload(channel, line string) {
	result := RE_DJHANDLER.FindAllStringSubmatch(line, -1)

	if len(result) == 1 {
		yid := result[0][2]
		go DownloadYID(yid)
		response := fmt.Sprintf("%s added to download queue", yid)
		irccon.Privmsg(channel, response)
	} else {
		fmt.Printf("no results found\n")
	}
}

func HandleNext(channel, line string) {
	fileName := MpdClient.Next()
	response := fmt.Sprintf("Now playing: %s", fileName)
	irccon.Privmsg(channel, response)
}

func HandleNowPlaying(channel, line string) {
	fileName := MpdClient.NowPlaying()
	response := fmt.Sprintf("Now playing: %s", fileName)
	irccon.Privmsg(channel, response)
}

func HandleRadioUrl(channel, line string) {
	response := fmt.Sprintf("Cant get enough of DJShuffle and Sjaak?? Sick of berms youtube links?? Listen to %s", Config.Bot.StreamURL)
	irccon.Privmsg(channel, response)
}

func ParsePrivmsg(e *irc.Event) {
	if len(e.Arguments) != 2 {
		return
	}

	channel := e.Arguments[0]
	line := e.Arguments[1]

	cmdResult := RE_CMD.FindAllStringSubmatch(line, -1)
	if len(cmdResult) != 1 {
		return
	}

	cmd := cmdResult[0][1]

	command, ok := isValidCommand(cmd)
	if !ok {
		return
	}

	switch command {
	case CMD_DJPLUS:
		HandleYidDownload(channel, line)
	case CMD_NEXT:
		HandleNext(channel, line)
	case CMD_PLAYING:
		HandleNowPlaying(channel, line)
	case CMD_RADIO:
		HandleRadioUrl(channel, line)
	}
}

func RunIrcBot() {
	server := fmt.Sprintf("%s:%d", Config.IRC.Server, Config.IRC.Port)

	irccon = irc.IRC(Config.IRC.Nickname, Config.IRC.Nickname)
	irccon.VerboseCallbackHandler = true
	irccon.Debug = false
	irccon.UseTLS = Config.IRC.UseTLS
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(Config.IRC.Channel) })
	irccon.AddCallback("366", func(e *irc.Event) {})
	irccon.AddCallback("PRIVMSG", ParsePrivmsg)
	err := irccon.Connect(server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.Loop()
}

func main() {
	var err error

	flag.Parse()

	Config, err = LoadConfig(*cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	MpdClient, err = NewMPD()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	chanName := stripChannel(Config.IRC.Channel)
	musicDir = fmt.Sprintf("%s/%s", Config.Youtube.BaseDir, chanName)

	pidFile := fmt.Sprintf("/var/musicbot/%s-%s.pid", Config.IRC.Nickname, chanName)
	logFile := fmt.Sprintf("/var/log/musicbot/%s-%s.log", Config.IRC.Nickname, chanName)

	if Config.IRC.Daemonize {
		ctx := daemon.Context{
			PidFileName: pidFile,
			PidFilePerm: 0644,
			LogFileName: logFile,
			LogFilePerm: 0640,
			WorkDir:     "/tmp",
			Umask:       022,
			Args:        []string{},
		}

		d, err := ctx.Reborn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to run as daemon: %v", err)
			os.Exit(1)
		}
		if d != nil {
			return
		}
		defer ctx.Release()
	}

	RunIrcBot()
}
