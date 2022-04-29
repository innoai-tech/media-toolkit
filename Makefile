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

web.fmt:
	 ./node_modules/.bin/prettier -w ./src

MTK = go run ./cmd/mtk

dep:
	go get -u -t ./pkg/...

serve: tidy
	$(MTK) -v1 serve -c .tmp/streams.json

fmt:
	goimports -w ./cmd/
	goimports -w ./pkg/

tidy:
	go mod tidy
