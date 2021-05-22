# ==== General configuration ====
# Bind address for source input
set("harbor.bind_addr", "0.0.0.0")

set("server.telnet.bind_addr", "0.0.0.0")
set("server.telnet.port", 2323)
set("server.telnet", true)
set("server.timeout", -1.0)

icy_source_pass = "development"
# mpd_source_pass = "mod_source"
live_source_pass = "live_source"

# === Functions ===


# ==== Input sources ====

# Silent track for use when the stream falls silent
emergency = mksafe(single("/etc/liquidsoap/silence.ogg"))

# Promo source once every hour
promo = amplify(1.2,
  delay(3600.,
    playlist(
      mode='random',
      "/etc/liquidsoap/promo",
      reload_mode="watch"
    )
  )
)

# List of jingles, to be played every N tracks
jingles = amplify(1.5,
  audio_to_stereo(
    playlist(
      mode='random',
      "/etc/liquidsoap/jingles",
      reload_mode="watch"
    )
  )
)

tracks = playlist(
    mode='random',
    "/music",
    reload_mode="watch"
)

requests = request.queue(id="request")

# Accept input from a live stream
live = strip_blank(
  audio_to_stereo(input.harbor("2600nl-live", port=8008, password=live_source_pass))
)

# ==== Configure radio station ====
# Mix in a jingle on the mpd stream every 3rd track
tracks = rotate(weights=[1,4], [jingles, tracks])

combined = fallback([requests, tracks])

# Mix in all different inputs and create the actual radio stream. The
# order in which all inputs are played is as follows:
# 1) Someone playing a live set via this radio station
# 2) A promo once every hour
# 3) The incoming mpd stream combined with jingles and talkshows
# 4) Silence
radio = fallback(track_sensitive=false, [live, promo, combined, emergency])

# ==== Output definitions ====
# Transcode 2600nl output to mp3 and forward it to a local icecast instance
output.icecast(
  %mp3(bitrate=192, stereo=true),
  mount="2600nl.mp3", icy_metadata="true",
  host="icecast", port=8000, password=icy_source_pass,
  radio
)

# Transcode 2600nl output to ogg and forward it to a local icecast instance
output.icecast(
  %vorbis.cbr(bitrate=192, channels=2),
  mount="2600nl.ogg", icy_metadata="true",
  host="icecast", port=8000, password=icy_source_pass,
  radio
)

## Avoid buffer overrruns
output.dummy(fallible=true, radio)
