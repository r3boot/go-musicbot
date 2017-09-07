TARGET = musicbot

BUILD_DIR = ./build

${TARGET}:
	[[ -d ${BUILD_DIR} ]] || mkdir -p ${BUILD_DIR}
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go
