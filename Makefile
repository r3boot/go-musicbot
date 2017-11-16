TARGET = musicbot
ID3TAG = id3tag

BUILD_DIR = ./build
PREFIX = /usr/local

all: ${TARGET} ${ID3TAG}

${TARGET}:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go

${ID3TAG}:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${ID3TAG} cmd/${ID3TAG}/${ID3TAG}.go

install:
	install -o root -m 0644 config/musicbot.yaml /etc/musicbot.yaml
	install -o root -m 0755 ${BUILD_DIR}/${ID3TAG} ${PREFIX}/bin/${ID3TAG}
	install -o root -m 0755 ${BUILD_DIR}/${ID3TAG} ${PREFIX}/bin/${ID3TAG}
	install -d -o root -g wheel -m 0755 webassets /usr/local/share/musicbot

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}"
