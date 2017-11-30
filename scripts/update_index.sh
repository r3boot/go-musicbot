#!/usr/bin/env bash

PLAYLIST_LOG='/var/log/icecast/playlist.log'

TEMPLATE='/usr/local/share/musicbot/templates/index.html'
OUTPUT='/usr/local/share/musicbot/index.html'

RADIO_2600NL_OGG_LISTENERS=$(grep '/2600nl.ogg' ${PLAYLIST_LOG} | grep -v ^Binary | tail -1 | cut -d\| -f3)
RADIO_2600NL_MP3_LISTENERS=$(grep '/2600nl.mp3' ${PLAYLIST_LOG} | grep -v ^Binary | tail -1 | cut -d\| -f3)
RADIO_2600NL_LISTENERS=$(expr ${RADIO_2600NL_OGG_LISTENERS:-0} + ${RADIO_2600NL_MP3_LISTENERS:-0})
RADIO_2600NL_NOWPLAYING="$(mpc status | head -1)"

RADIO_TAPES_OGG_LISTENERS=$(grep '/tapes.ogg' ${PLAYLIST_LOG} | grep -v ^Binary | tail -1 | cut -d\| -f3)
RADIO_TAPES_MP3_LISTENERS=$(grep '/tapes.mp3' ${PLAYLIST_LOG} | grep -v ^Binary | tail -1 | cut -d\| -f3)
RADIO_TAPES_LISTENERS=$(expr ${RADIO_TAPES_OGG_LISTENERS:-0} + ${RADIO_TAPES_MP3_LISTENERS:-0})
RADIO_TAPES_NOWPLAYING="$(echo -e '/tapes(dot)mp3.metadata\nquit\n' | nc localhost 6009 | grep ^filename= | tail -1 | cut -d\" -f2 | cut -d/ -f4,5 | sed -e 's,.mp3,,g' -e 's,.MP3,,g' -e 's,.m4a,,g')"

cat "${TEMPLATE}" | sed \
  -e "s,%2600NL_NOWPLAYING%,${RADIO_2600NL_NOWPLAYING},g" \
  -e "s,%2600NL_NUMLISTENERS%,${RADIO_2600NL_LISTENERS},g" \
  -e "s,%TAPES_NOWPLAYING%,${RADIO_TAPES_NOWPLAYING},g" \
  -e "s,%TAPES_NUMLISTENERS%,${RADIO_TAPES_LISTENERS},g" \
  > ${OUTPUT}.new && mv ${OUTPUT}.new ${OUTPUT}
