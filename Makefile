TARGET = musicbot
CLI = ${TARGET}-cli

BUILD_DIR = ./build
PREFIX = /usr/local

all: ${TARGET} ${CLI}

deps:
	go get -v ./...

${TARGET}:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go

${CLI}:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${CLI} cmd/${CLI}/${CLI}.go

install:
	install -o root -m 0644 config/musicbot.yaml /etc/musicbot.yaml
	install -o root -m 0755 ${BUILD_DIR}/${TARGET} ${PREFIX}/bin/${TARGET}
	install -o root -m 0755 ${BUILD_DIR}/${CLI} ${PREFIX}/bin/${CLI}
	install -d -o root -g wheel -m 0755 webassets /usr/local/share/musicbot

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}"
