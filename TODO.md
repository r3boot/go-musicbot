# musicbot
- Add possibility to send reply via /q
- Add feature to create a favourites list based on rating
- Make control-char actually configurable
- Add a negative rating to the person who submitted a yid when rating == 0
- Make playlist download possible using ID instead of URL
- Add possibility for an ACL on various commands
- Bind !dj+ command to username in some database
- Make sure we report proper duration/elapsed times
- Make ratings drift back to the starting point over time
- Add duplicate check
- Rework favourites/toptracks playlist into mp3 priorities
- Rework ratings to use track field
- Keep playlist prio + play queue in sync
- Clear playqueue on startup

# liquidsoap
- Make sure its clear which program is playing
- Make sure that !tune/!boo/!next redirects to liquidsoap if required
- Add leading + following jingle to favourites + toptracks
- Checkout http://savonet.sourceforge.net/doc-svn/radiopi.html

# webui/api
- Add attribution for next function in webapi
- Add scroller for keeping track of webapi commands
- Auto reload when websocket is closed
- Make api process dedicated per stream
- Send back valid JSON when TA search is less then 3 chars

# Future
- Download all videos, investigate video stream
- Investigate blockchain technology to moderate webui controls
- Look into possibility of writing a mobile app

# Stream
- Embed metadata into stream
- Query metadata for id3 tags
