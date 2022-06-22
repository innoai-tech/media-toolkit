package main

import (
	"strings"

	"dagger.io/dagger"
	"dagger.io/dagger/core"

	"github.com/innoai-tech/runtime/cuepkg/tool"
	"github.com/innoai-tech/runtime/cuepkg/node"
	"github.com/innoai-tech/runtime/cuepkg/golang"
	"github.com/innoai-tech/runtime/cuepkg/imagetool"
)

dagger.#Plan

client: env: {
	VERSION: string | *"dev"
	GIT_SHA: string | *""
	GIT_REF: string | *""

	GOPROXY:   string | *""
	GOPRIVATE: string | *""
	GOSUMDB:   string | *""

	GH_USERNAME: string | *""
	GH_PASSWORD: dagger.#Secret

	LINUX_MIRROR:                  string | *""
	CONTAINER_REGISTRY_PULL_PROXY: string | *""
}

actions: version: tool.#ResolveVersion & {
	ref:     "\(client.env.GIT_REF)"
	version: "\(client.env.VERSION)"
}

auths: {
	"ghcr.io": {
		username: "\(client.env.GH_USERNAME)"
		secret:   client.env.GH_PASSWORD
	}
}

mirror: {
	linux: "\(client.env.LINUX_MIRROR)"
	pull:  "\(client.env.CONTAINER_REGISTRY_PULL_PROXY)"
}

actions: liveplayer: node.#ViteProject & {
	source: {
		path: "."
		include: [
			"webapp/",
			"package.json",
			"pnpm-lock.yaml",
			"tsconfig.json",
			"vite.config.ts",
		]
	}

	env: APP: "live-player"

	build: {
		pre: [
			"pnpm install",
		]
		image: {
			"mirror": mirror
			"auths":  auths
			steps: [
				node.#ConfigPrivateRegistry & {
					scope: "@innoai-tech"
					host:  "npm.pkg.github.com"
					token: client.env.GH_PASSWORD
				},
				imagetool.#Script & {
					run: [
						"npm i -g pnpm",
					]
				},
			]
		}
	}
}

dependences: {
	"ghcr.io/innoai-tech/ffmpeg": "5"
}

actions: go: golang.#Project & {
	source: {
		path: "."
		include: [
			"cmd/",
			"pkg/",
			"internal/",
			"go.mod",
			"go.sum",
		]
	}

	version:  "\(actions.version.output)"
	revision: "\(client.env.GIT_SHA)"

	// when disable cross-gcc will be installed
	isolate: false
	cgo:     true

	goos: ["linux"]
	goarch: ["amd64", "arm64"]
	main: "./cmd/mtk"
	ldflags: [
		"-s -w",
		"-X \(go.module)/pkg/version.Version=\(go.version)",
		"-X \(go.module)/pkg/version.Revision=\(go.revision)",
	]

	mounts: {
		webui: core.#Mount & {
			contents: actions.liveplayer.build.output.rootfs
			source:   "/output"
			dest:     "\(go.workdir)/internal/liveplayer/dist"
		}
	}

	env: {
		GOPROXY:   client.env.GOPROXY
		GOPRIVATE: client.env.GOPRIVATE
		GOSUMDB:   client.env.GOSUMDB
	}

	build: {
		pre: [
			"go mod download",
		]
		image: {
			steps: [
				imagetool.#ImageDep & {
					// for cross compile need to load .so for all platforms
					"platforms": [ for arch in goarch {
						"linux/\(arch)"
					}]
					"dependences": dependences
					"auths":       auths
					"mirror":      mirror
				},
			]
		}
	}

	ship: {
		name: "\(strings.Replace(go.module, "github.com/", "ghcr.io/", -1))"

		from: "gcr.io/distroless/cc-debian11:debug"
		steps: [
			imagetool.#ImageDep & {
				"dependences": dependences
				"auths":       auths
				"mirror":      mirror
			},
		]

		config: {
			cmd: ["serve"]
		}
	}

	build: image: {
		"auths":  auths
		"mirror": mirror
	}

	ship: {
		"auths":  auths
		"mirror": mirror
	}

	devkit: load: host: client.network."unix:///var/run/docker.sock".connect
	ship: load: host:   client.network."unix:///var/run/docker.sock".connect
}

client: filesystem: "build/output": write: contents: actions.go.archive.output
client: network: "unix:///var/run/docker.sock": connect: dagger.#Socket
