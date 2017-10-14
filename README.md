# Introduction
Go-musicbot is an IRC bot which responds which  is built around
downloading youtube videos, converting them to the MP3 format (using
youtube-dl) and adding these MP3's to an MPD stream. It also supports
basic functionality for MPD.

# Features
* Download youtube videos. A log is kept of all already downloaded files
* Control Next, NowPlaying and UpdateDB functionality for MPD
* Display a link to the associated radio stream
* Maintain a rating for a song.
* Remove song when rating drops to 0

# Commands
Downloading a new link:
~~~~
<@r3boot> !dj+ DBHipNYuAZk
< IrcBot> DBHipNYuAZk added to download queue
~~~~

Downloading a playlist:
~~~~
<@r3boot> !djlist https://www.youtube.com/watch?v=RkENy3t_fgQ&index=15&list=RDGMEMYH9CUrFO7CfLJpaD7UR85wVMhJpxwMtWy_A
< IrcBot> Added playlist to download queue
~~~~

Press play on mpd
~~~~
<@r3boot> !play
< IrcBot> Now playing: Mark Instinct - Bad Seed [HQ] [DOWNLOAD]-SYbGij3o0ec.mp3

~~~~

Displaying current song:
~~~~
<@r3boot> !np
< IrcBot> Now playing: Fear Factory - T-1000 (H-K) (duration: 4m34s; rating: 5/10)
~~~~

Switching to next song:
~~~~
<@r3boot> !next
< IrcBot> Now playing: Massive Attack - Special Cases (duration: 5m16s; rating: 5/10)
~~~~

Displaying the URL for the radio:
~~~~
<@r3boot> !radio
< IrcBot> Fed up with the youtube links of berm? Listen to http://radio.as65342.net:8000/2600nl.ogg.m3u
~~~~

Increase the rating for the currently playing song
~~~~
<@r3boot> !tune
< IrcBot> Rating for Dioxyde - Torschlüsspanik-_DhqHr6IY78.mp3 is 7/10 .. Party on!!!!

~~~~

Decrease the rating for the currently playing song
~~~~
<@r3boot> !boo
< IrcBot> Rating for Nirvana - Rape Me (Legendado)-n8aJG1f815Q.mp3 is 4/10 .. BOOO!!!!
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

# Ratings
Every song can have a rating of 1..10. The default rating is 5. As soon
as the rating of a song drops below 1, it will be removed from the
playlist.
