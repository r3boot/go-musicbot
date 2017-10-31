TARGET = musicbot

BUILD_DIR = ./build
PREFIX = /usr/local

all: ${TARGET} api

${TARGET}:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go

api:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${TARGET}-api cmd/${TARGET}-api/${TARGET}-api.go

install:
	install -o root -m 0644 config/musicbot.yaml /etc/musicbot.yaml
	install -o root -m 0755 ${BUILD_DIR}/${TARGET} ${PREFIX}/bin/${TARGET}
	install -o root -m 0755 ${BUILD_DIR}/${TARGET}-api ${PREFIX}/bin/${TARGET}-api
	install -d -o root -g root -m 0755 webassets /usr/local/share/musicbot

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}"
