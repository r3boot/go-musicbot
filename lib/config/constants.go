package config

type ServicesConfig struct {
	IrcBotEnabled bool `yaml:"ircbot"`
	APIEnabled    bool `yaml:"api"`
	WebUIEnabled  bool `yaml:"webui"`
}

type IrcConfig struct {
	Nickname  string `yaml:"nickname"`
	Server    string `yaml:"server"`
	Port      int    `yaml:"port"`
	Channel   string `yaml:"channel"`
	UseTLS    bool   `yaml:"tls"`
	VerifyTLS bool   `yaml:"tls_verify"`
	Daemonize bool   `yaml:"daemonize"`
	Debug     bool   `yaml:"debug"`
}

type BotConfig struct {
	CommandChar   string   `yaml:"command_character"`
	ValidCommands []string `yaml:"valid_commands"`
	StreamURL     string   `yaml:"stream_url"`
	RadioMsgs     []string `yaml:"radio_messages"`
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

type ApiConfig struct {
	Address   string `yaml:"address"`
	Port      string `yaml:"port"`
	Title     string `yaml:"title"`
	StreamURL string `yaml:"stream_url"`
	Assets    string `yaml:"assets"`
}

type MusicBotConfig struct {
	Services ServicesConfig `yaml:"services"`
	IRC      IrcConfig      `yaml:"irc"`
	Bot      BotConfig      `yaml:"bot"`
	Youtube  YoutubeConfig  `yaml:"youtube"`
	MPD      MpdConfig      `yaml:"mpd"`
	Api      ApiConfig      `yaml:"api"`
}
