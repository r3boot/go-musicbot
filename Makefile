TARGETS   = musicbot musicbotctl musicbot-webui
UTILS     = embed-assets
NAMESPACE = as65342

BUILD_DIR   = ./build
SCRIPT_DIR  = ./scripts
DOC_DIR     = ./doc
DOCKER_DIR  = ./docker
COMPOSE_DIR = ./compose
LIB_DIR     = ./lib

LATEST_TAG = $(NAMESPACE)/musicbot:latest
BUILD_TAG  = $(NAMESPACE)/musicbot-build:latest

all: $(TARGETS)

$(BUILD_DIR):
	mkdir -p "$(BUILD_DIR)"

validate:
	swagger validate swagger.yaml

generate: validate
	swagger generate server -f swagger.yaml -s ./lib/apiserver --with-expand --exclude-main
	swagger generate client -f swagger.yaml -c ./lib/apiclient --with-expand

$(UTILS): $(BUILD_DIR)
	go build -v -o $</$@ scripts/$@.go

generate_assets: $(UTILS)
	$(BUILD_DIR)/embed-assets > $(LIB_DIR)/webui/assets.go

$(TARGETS): $(BUILD_DIR)
	go build -v -o $</$@ cmd/$@/main.go

build_container:
	docker image rm $(BUILD_TAG) || true
	docker build --no-cache --rm -t $(BUILD_TAG) -f $(DOCKER_DIR)/Dockerfile.build $(DOCKER_DIR)/

container:
	install -m 0755 $(BUILD_DIR)/musicbot $(DOCKER_DIR)/files/musicbot
	docker build --no-cache --rm -t $(LATEST_TAG) -f $(DOCKER_DIR)/Dockerfile $(DOCKER_DIR)/

compile:
	docker run -it --rm \
		-v ${GOPATH}:/go \
		$(BUILD_TAG)

graphs:
	cd $(DOC_DIR) ; make -f Makefile all

test: generate compile container
	cd $(COMPOSE_DIR) ; docker-compose up

clean:
	[[ -d "$(BUILD_DIR)" ]] && rm -rf "$(BUILD_DIR)" || true
