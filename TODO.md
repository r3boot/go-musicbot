# musicbot
- Add possibility to send reply via /q
- Make control-char actually configurable
- Add a negative rating to the person who submitted a yid when rating == 0
- Make playlist download possible using ID instead of URL
- Add possibility for an ACL on various commands
- Bind !dj+ command to username in some database
- Make ratings drift back to the starting point over time
- Add duplicate check
- Keep playlist prio + play queue in sync
- Check running track @ enqueue
- Download non-existing track on !request
- Add option to de-queue track
- [bug] unable to start bot when mpd is not running
- Automatically add DJShuffle tracks to playlist

# liquidsoap
- Make sure its clear which program is playing
- Make sure that !tune/!boo/!next redirects to liquidsoap if required
- Add leading + following jingle to favourites + toptracks
- Checkout http://savonet.sourceforge.net/doc-svn/radiopi.html

# webui/api
- Add auditing for various actions
- Add attribution for functions in webapi
- Make api process dedicated per stream
- Fix height calculation of pagination

# Stream
- Embed metadata into stream
- Query metadata for id3 tags

# Future
- Download all videos, investigate video stream
- Investigate blockchain technology to moderate webui controls
- Look into possibility of writing a mobile app
