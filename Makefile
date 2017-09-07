TARGET = musicbot

BUILD_DIR = ./build
PREFIX = /usr/local

${TARGET}:
	[[ -d ${BUILD_DIR} ]] || mkdir -p ${BUILD_DIR}
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go

install:
	install -o root -m 0755 ${BUILD_DIR}/${TARGET} ${PREFIX}/bin/${TARGET}
