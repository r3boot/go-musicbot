# musicbot
- Add possibility to send reply via /q
- Add feature to create a favourites list based on rating
- Make control-char actually configurable
- Add a negative rating to the person who submitted a yid when rating == 0
- Make sure rating isnt overwritten when downloading double item
- Make playlist download possible using ID instead of URL
- Add possibility for an ACL on various commands

# webui/api
- Add attribution for next function in webapi
- Add scroller for keeping track of webapi commands
- Add list of currently enqueued songs
- Auto reload when websocket is closed
- Make api process dedicated per stream

# Future
- Download all videos, investigate video stream
- Investigate blockchain technology to moderate webui controls
- Build notification bridge between musicbot and musicbot-api
- Improve queue handling (make users more aware of queue items)
