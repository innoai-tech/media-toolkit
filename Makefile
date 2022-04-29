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
	docker pull ghcr.io/innoai-tech/media-toolkit:main
	docker run \
		-it \
		-v=$(PWD)/.tmp:/.tmp \
		-p=777:777 \
		ghcr.io/innoai-tech/media-toolkit:main -v1 serve -c .tmp/streams.json

fmt:
	goimports -w ./cmd/
	goimports -w ./pkg/

cue.dep:
	cuem get -u ./...

tidy:
	go mod tidy

eval:
	cuem eval -o components.yaml ./cuepkg/mtk

import.debug:
	cuem eval -o local-dev.yaml ./.tmp/local-dev.cue > ./.tmp/local-dev.yaml
	kubepkg import -i=http://local-dev.office:36060 --incremental ./.tmp/local-dev.yaml