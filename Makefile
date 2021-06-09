TARGETS   = musicbot musicbotctl musicbot-webui musicbot-ircbot
UTILS     = embed-assets
NAMESPACE = as65342

BUILD_DIR   = ./build
SCRIPT_DIR  = ./scripts
DOC_DIR     = ./doc
DOCKER_DIR  = ${PWD}/docker
COMPOSE_DIR = ./compose
LIB_DIR     = ./lib

LATEST_TAG = $(NAMESPACE)/musicbot:latest
BUILD_TAG  = $(NAMESPACE)/musicbot-build:latest

MUSICBOT_BUILD_TAG = $(NAMESPACE)/musicbot-build:latest
MUSICBOT_WEBUI_BUILD_TAG = $(NAMESPACE)/musicbot-webui-build:latest
MUSICBOT_IRCBOT_BUILD_TAG = $(NAMESPACE)/musicbot-ircbot-build:latest

MUSICBOT_LATEST_TAG = $(NAMESPACE)/musicbot:latest
MUSICBOT_WEBUI_LATEST_TAG = $(NAMESPACE)/musicbot-webui:latest
MUSICBOT_IRCBOT_LATEST_TAG = $(NAMESPACE)/musicbot-ircbot:latest

all: $(TARGETS)

release: compile containers push

$(BUILD_DIR):
	mkdir -p "$(BUILD_DIR)"

validate:
	swagger validate swagger.yaml

generate: validate
	swagger generate server -f swagger.yaml -s ./lib/apiserver --with-expand --exclude-main --skip-models
	swagger generate client -f swagger.yaml -c ./lib/apiclient --with-expand --skip-models

$(UTILS): $(BUILD_DIR)
	go build -v -o $</$@ scripts/$@.go

generate_assets: $(UTILS)
	$(BUILD_DIR)/embed-assets > $(LIB_DIR)/webui/assets.go

$(TARGETS): $(BUILD_DIR)
	go build -v -o $</$@ cmd/$@/main.go

compile: build_container build_musicbot build_musicbot_webui build_musicbot_ircbot

build_container:
	docker image rm $(BUILD_TAG) || true
	docker build --no-cache --rm -t $(BUILD_TAG) -f $(DOCKER_DIR)/Dockerfile.build \
		$(DOCKER_DIR)/

build_musicbot:
	docker run -it --rm -v ${GOPATH}:/go -v $(DOCKER_DIR)/musicbot/files/build.sh:/build.sh $(BUILD_TAG)

build_musicbot_webui:
	docker run -it --rm -v ${GOPATH}:/go -v $(DOCKER_DIR)/musicbot-webui/files/build.sh:/build.sh $(BUILD_TAG)

build_musicbot_ircbot:
	docker run -it --rm -v ${GOPATH}:/go -v $(DOCKER_DIR)/musicbot-ircbot/files/build.sh:/build.sh $(BUILD_TAG)

containers: container_musicbot container_musicbot_webui container_musicbot_ircbot

container_musicbot:
	install -m 0755 $(BUILD_DIR)/musicbot $(DOCKER_DIR)/musicbot/files/musicbot
	docker build --no-cache --rm -t $(MUSICBOT_LATEST_TAG) -f $(DOCKER_DIR)/musicbot/Dockerfile \
		$(DOCKER_DIR)/musicbot/

container_musicbot_webui:
	install -m 0755 $(BUILD_DIR)/musicbot-webui $(DOCKER_DIR)/musicbot-webui/files/musicbot-webui
	docker build --no-cache --rm -t $(MUSICBOT_WEBUI_LATEST_TAG) -f $(DOCKER_DIR)/musicbot-webui/Dockerfile \
		$(DOCKER_DIR)/musicbot-webui/

container_musicbot_ircbot:
	install -m 0755 $(BUILD_DIR)/musicbot-ircbot $(DOCKER_DIR)/musicbot-ircbot/files/musicbot-ircbot
	docker build --no-cache --rm -t $(MUSICBOT_IRCBOT_LATEST_TAG) -f $(DOCKER_DIR)/musicbot-ircbot/Dockerfile \
		$(DOCKER_DIR)/musicbot-ircbot/

push: push_musicbot push_musicbot_webui push_musicbot_ircbot

push_musicbot:
	docker push $(MUSICBOT_LATEST_TAG)

push_musicbot_webui:
	docker push $(MUSICBOT_WEBUI_LATEST_TAG)

push_musicbot_ircbot:
	docker push $(MUSICBOT_IRCBOT_LATEST_TAG)

graphs:
	cd $(DOC_DIR) ; make -f Makefile all

test: clean generate compile containers
	cd $(COMPOSE_DIR) ; docker compose up

clean:
	[[ -d "$(BUILD_DIR)" ]] && rm -rf "$(BUILD_DIR)" || true
