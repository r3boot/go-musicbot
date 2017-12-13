#!/usr/bin/env bash

BASEDIR='/music/2600nl'

/usr/local/bin/mbfixtags "${BASEDIR}"
/usr/bin/mpc update
sleep 30
/usr/bin/systemctl restart musicbot-2600nl