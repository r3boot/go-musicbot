#!/usr/bin/env bash

if [[ ${#} -eq 1 ]]; then
  FNAME="${1}"

  id3v2 -l "${FNAME}" | grep -q "^TLEN"
  if [[ ${?} -eq 0 ]]; then
    echo "TLEN already set"
    exit 0
  fi

  DURATION="$(mp3info -p "%S" "${FNAME}")"
  id3v2 --TLEN ${DURATION} "${FNAME}"
else
  ls /music/2600nl/*.mp3 | while read FNAME; do
    echo "${FNAME}"
    id3v2 -l "${FNAME}" | grep -q "^TLEN"
    if [[ ${?} -eq 0 ]]; then
      continue
    fi

    DURATION="$(mp3info -p "%S" "${FNAME}")"
    id3v2 --TLEN ${DURATION} "${FNAME}"
  done

  mpc update
fi