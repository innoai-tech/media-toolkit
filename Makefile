export GIT_SHA ?= $(shell git rev-parse HEAD)
export GIT_REF ?= HEAD

debug:
	cd ../../dagger/dagger && go install ./cmd/dagger
	dagger do build web
.PHONY: debug

ship:
	dagger do webapp build
	dagger do go ship pushx

dep.web:
	 pnpm install

dev.web:
	cd webapp/live-player && pnpm run dev

build.web:
	pnpm exec turbo run build --force

lint.web:
	pnpm exec turbo run lint --force

MTK = CGO_ENABLED=1 go run ./cmd/mtk

dep:
	go get -u -t ./pkg/...

serve:
	$(MTK) -v1 serve -c .tmp/streams.json

serve.debug:
	docker run \
		-it \
		-v=$(PWD)/.tmp:/.tmp \
		-p=777:777 \
		ghcr.io/innoai-tech/media-toolkit@sha256:b3d49a06bc5d9ecf1670d440429169669bb6c1a76ab0ae4d3a507f8a4deb73e9 -v1 serve -c .tmp/streams.json

devkit:
	dagger do go devkit load/linux/arm64

dev:
	docker run \
		-it \
		-v=$(shell go env GOMODCACHE):/go/src/mod \
		-v=$(PWD):/go/src \
		-w=/go/src \
		-p=777:777 \
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

local.rtp:
	go run ./cmd/localasrtp 0.0.0.0:577