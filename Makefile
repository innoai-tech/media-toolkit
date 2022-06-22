export GIT_SHA ?= $(shell git rev-parse HEAD)
export GIT_REF ?= HEAD

DAGGER = dagger --log-format=plain -p ./

debug:
	cd ../../dagger/dagger && go install ./cmd/dagger
	dagger do build web
.PHONY: debug

build:
	$(DAGGER) do build
.PHONY: build

push:
	$(DAGGER) do push
.PHONY: push

web.dep:
	 pnpm install

web.dev:
	./node_modules/.bin/vite --config=vite.config.ts

web.build:
	./node_modules/.bin/vite build  --mode=production --config=vite.config.ts

web.fmt:
	 ./node_modules/.bin/prettier -w ./webapp

MTK = go run ./cmd/mtk

dep:
	go get -u -t ./pkg/...

serve: tidy
	$(MTK) -v1 serve -c .tmp/streams.json

serve.debug:
	docker run \
		-it \
		-v=$(PWD)/.tmp:/.tmp \
		-p=777:777 \
		ghcr.io/innoai-tech/media-toolkit:dev -v1 serve -c .tmp/streams.json

devkit:
	dagger do go devkit load arm64

dev:
	docker run \
		-it \
		-v=$(shell go env GOMODCACHE):/go/src/mod \
		-v=$(PWD):/go/src \
		-w=/go/src \
		github.com/innoai-tech/media-toolkit:devkit-arm64

xx.amd64:
	CC=x86_64-linux-gnu-gcc \
	CXX=x86_64-linux-gnu-g++ \
	CGO_ENABLED=1 \
	GOARCH=amd64  \
	go build \
		-ldflags="-linkmode=external" \
		-o build/mtk-linux-amd64 \
		./cmd/mtk

xx.arm64:
	CC=aarch64-linux-gnu-gcc \
	CXX=aarch64-linux-gnu-g++ \
	CGO_ENABLED=1 \
    GOARCH=arm64  \
	go build \
		-ldflags="-linkmode=external" \
		-o build/mtk-linux-arm64 \
		./cmd/mtk

fmt:
	goimports -w ./cmd/
	goimports -w ./pkg/

cue.dep:
	cuem get -u ./...

tidy:
	go mod tidy

eval:
	cuem eval -o components.yaml ./cuepkg/mtk

export.kubepkg:
	cuem eval -o local-dev.yaml ./.tmp/local-dev.cue > .tmp/local-dev.yaml

import.debug: export.kubepkg
	kubepkg import -i=http://local-dev.office:36060 --incremental .tmp/local-dev.yaml