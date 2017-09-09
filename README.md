# Introduction
Go-musicbot is an IRC bot which responds which  is built around
downloading youtube videos, converting them to the MP3 format (using
youtube-dl) and adding these MP3's to an MPD stream. It also supports
basic functionality for MPD.

# Features
* Download youtube videos. A log is kept of all already downloaded files
* Control Next, NowPlaying and UpdateDB functionality for MPD
* Display a link to the associated radio stream

# Commands
Downloading a new link:
~~~~
<@r3boot> !dj+ DBHipNYuAZk
< IrcBot> DBHipNYuAZk added to download queue
~~~~

Displaying current song:
~~~~
<@r3boot> !np
< IrcBot> Now playing: Funana So Cu Pe-yjfVuVvz6aA.mp3
~~~~

Switching to next song:
~~~~
<@r3boot> !next
< IrcBot> Now playing: Wale Ft. Tiara Thomas -Bad (Official Video)-0TIIu9CERgI.mp3
~~~~

Displaying the URL for the radio:
~~~~
<@r3boot> !radio
< IrcBot> Fed up with the youtube links of berm? Listen to http://radio.as65342.net:8000/2600nl.ogg.m3u
~~~~

# Installation
First, fetch the code
~~~~
go get -v github.com/r3boot/go-musicbot
~~~~

Then, navigate to the repo and build & install the bot
~~~~
cd ${GOPATH}/src/github.com/r3boot/go-musicbot
make && sudo make install
~~~~

# Configuration
Please see /etc/musicbot.yaml once the bot has been installed. In this
file you will need to configure your IRC and MPD details. To connect
the bot to multiple networks / channels, copy the configuration file
to a new name and edit the configuration. Next run multiple instances
of this bot.

# Running
Run the following command (as a non-root user):
~~~~
/usr/local/bin/musicbot -f /etc/musicbot.yaml
~~~~