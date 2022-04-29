web.dev:
	pnpx vite --config=vite.config.ts

web.pkg:
	pnpx vite --config=vite.config.ts build --mode=production

web.fmt:
	pnpx prettier -w ./web/src

MTK = go run ./cmd/mtk

live:
	$(MTK) -v1 live play rtmp://xxxx

fmt:
	goimports -w .

tidy:
	go mod tidy

npm.install:
	pnpm install

cp: npm.install
