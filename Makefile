TARGET = musicbot

BUILD_DIR = ./build
PREFIX = /usr/local

all: ${TARGET} api

api: validate_api generate_api go-${TARGET}-server

validate_api:
	swagger validate musicapi.yaml

generate_api:
	swagger generate server -f musicapi.yaml
	#go get -u -f ./...

${TARGET}:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/${TARGET} cmd/${TARGET}/${TARGET}.go

go-${TARGET}-server:
	[[ -d "${BUILD_DIR}" ]] || mkdir -p "${BUILD_DIR}"
	go build -v -o ${BUILD_DIR}/go-${TARGET}-server cmd/go-${TARGET}-server/main.go

install:
	install -o root -m 0644 config/musicbot.yaml /etc/musicbot.yaml
	install -o root -m 0755 ${BUILD_DIR}/${TARGET} ${PREFIX}/bin/${TARGET}

clean:
	[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}"
	[[ -d "cmd/go-${TARGET}-server" ]] && rm -rf "cmd/go-${TARGET}-server"
	[[ -d "models" ]] && rm -rf "models"
	[[ -d "restapi/operations" ]] && rm -rf "restapi/operations"
	[[ -f "restapi/server.go" ]] && rm -f "restapi/server.go"
	[[ -f "restapi/doc.go" ]] && rm -f "restapi/doc.go"
	[[ -f "restapi/embedded_spec.go" ]] && rm -f "restapi/embedded_spec.go"
