FONTAWESOME_VERSION=6.1.1
TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
GIT_TAG := $(shell git describe --abbrev=0 --tags $(TAG_COMMIT) 2>/dev/null || true)
VERSION := $(shell echo $(GIT_TAG) | sed 's/^.\{1\}//')
DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

.PHONY:build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${GIT_TAG}" -o build/gosh ./cmd/gosh
	chmod u+x build/gosh

publish: build-docker
	docker push stenehall/gosh:$(VERSION)
	docker tag stenehall/gosh:$(VERSION) stenehall/gosh:latest
	docker push stenehall/gosh:latest

build-docker:
	docker build -t stenehall/gosh:$(VERSION) --no-cache --build-arg BUILD_DATE=$(DATE) --build-arg BUILD_VERSION=$(VERSION) .

docker-up:
	docker compose up

fetch-fontawesome: clean
	mkdir -p assets/fontawesome/css
	wget https://use.fontawesome.com/releases/v$(FONTAWESOME_VERSION)/fontawesome-free-$(FONTAWESOME_VERSION)-web.zip
	unzip -q -d assets fontawesome-free-$(FONTAWESOME_VERSION)-web.zip
	mv assets/fontawesome-free-$(FONTAWESOME_VERSION)-web/css/all.min.css assets/fontawesome/css/all.min.css
	mv assets/fontawesome-free-$(FONTAWESOME_VERSION)-web/webfonts assets/fontawesome/webfonts
	mv assets/fontawesome-free-$(FONTAWESOME_VERSION)-web/svgs/regular/snowflake.svg assets/favicon.svg
	rm -rf assets/fontawesome/webfonts/*.ttf
	rm -rf assets/fontawesome-free-$(FONTAWESOME_VERSION)-web
	rm -rf fontawesome-free-$(FONTAWESOME_VERSION)-web.zip

test-docker: clean
	docker build -t gosh .
	mkdir -p test/favicons
	cp configs-example/config.yml ./test
	cp docker-compose.yml ./test
	cd test && docker compose up


clean:
	rm -rf assets/fontawesome
	rm -rf favicons
	rm -rf build
	rm -rf test
