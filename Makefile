TARGET = musicbot
CLI = ${TARGET}-cli
MBFIXTAGS = mbfixtags

BUILD_DIR = ./build
CONFIG_DIR = ./config
PREFIX = /usr/local

all: ${TARGET} ${CLI}

deps:
	go get -v ./...

${BUILD_DIR}:
	mkdir -p "${BUILD_DIR}"

${TARGET}: ${BUILD_DIR}
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go

${CLI}: ${BUILD_DIR}
	go build -v -o ${BUILD_DIR}/${CLI} cmd/${CLI}/${CLI}.go

${MBFIXTAGS}: ${BUILD_DIR}
	go build -v -o ${BUILD_DIR}/${MBFIXTAGS} cmd/${MBFIXTAGS}/${MBFIXTAGS}.go

clean_mpd: stop_mpd
	rm -f ${CONFIG_DIR}/mpd.db
	rm -f ${CONFIG_DIR}/mpdstate

start_mpd: clean_mpd
	rm -f ${CONFIG_DIR}/mpd.db
	mpd ${CONFIG_DIR}/mpd.conf
	sleep 1
	mpc update

stop_mpd:
	mpd ${CONFIG_DIR}/mpd.conf --kill | true

install:
	strip -v ${BUILD_DIR}/${TARGET}
	strip -v ${BUILD_DIR}/${CLI}
	strip -v ${BUILD_DIR}/${MBFIXTAGS}
	install -o root -m 0644 config/musicbot.yaml /etc/musicbot.yaml
	install -o root -m 0755 ${BUILD_DIR}/${TARGET} ${PREFIX}/bin/${TARGET}
	install -o root -m 0755 ${BUILD_DIR}/${CLI} ${PREFIX}/bin/${CLI}
	install -o root -m 0755 ${BUILD_DIR}/${MBFIXTAGS} ${PREFIX}/bin/${MBFIXTAGS}
	install -d -o root -g wheel -m 0755 ${PREFIX}/share/musicbot
	cp -Rp webassets/* ${PREFIX}/share/musicbot/
	install -o root -m 0755 scripts/update_index.sh ${PREFIX}/bin/update_index.sh
	install -o root -m 0755 scripts/fix_duration.sh ${PREFIX}/bin/fix_duration.sh
	install -o root -m 0755 scripts/fix_id3tags.sh ${PREFIX}/bin/fix_id3tags.sh

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}"
