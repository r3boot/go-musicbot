#!/usr/bin/env bash

BASEDIR='/music/2600nl'

ls "${BASEDIR}"/*.mp3 | while read FULLPATH; do
  FNAME="$(basename "${FULLPATH}")"
  FNAME="${FNAME::len-16}"

  echo "${FULLPATH}"
  echo "${FNAME}" | grep -q '-'
  if [[ ${?} -eq 0 ]]; then
    ARTIST="$(echo "${FNAME}" | cut -d '-' -f1 | sed -e 's,\ +$,,g')"
    TITLE="$(echo "${FNAME}" | cut -d '-' -f2,3,4,5 | sed -e 's,^\ +,,g')"
    id3v2 -a "${ARTIST}" -t "${TITLE}" "${FULLPATH}"
  else
    TITLE="$(echo "${FNAME}" | awk -F '-' '{print $1}')"
    id3v2 -t "${TITLE}" "${FULLPATH}"
  fi
done