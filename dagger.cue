package main

import (
	"dagger.io/dagger"
	"dagger.io/dagger/core"
	"universe.dagger.io/docker"

	"github.com/innoai-tech/runtime/cuepkg/tool"
	"github.com/innoai-tech/runtime/cuepkg/debian"
	"github.com/innoai-tech/runtime/cuepkg/golang"
	"github.com/innoai-tech/runtime/cuepkg/node"
)

dagger.#Plan & {
	client: {
		env: {
			VERSION: string | *"dev"
			GIT_SHA: string | *""
			GIT_REF: string | *""

			GOPROXY:   string | *""
			GOPRIVATE: string | *""
			GOSUMDB:   string | *""

			GH_USERNAME: string | *""
			GH_PASSWORD: dagger.#Secret

			LINUX_MIRROR: string | *""
		}

		filesystem: "./build/output": write: contents: actions.export.output
	}

	actions: {
		_goSource: core.#Source & {
			path: "."
			include: [
				"cmd/",
				"internal/",
				"pkg/",
				"go.mod",
				"go.sum",
			]
		}

		_webappSource: core.#Source & {
			path: "."
			include: [
				"webapp/",
				"package.json",
				"pnpm-lock.yaml",
				"tsconfig.json",
				"vite.config.ts",
			]
		}

		_imageName: "ghcr.io/innoai-tech/media-toolkit"
		_version:   (tool.#ResolveVersion & {ref: client.env.GIT_REF, version: "\(client.env.VERSION)"}).output
		_tag:       _version

		info: golang.#Info & {
			source: _goSource.output
		}

		_archs: ["amd64", "arm64"]

		_deps: {
			// https://packages.debian.org/bullseye/libavutil-dev
			"libavutil-dev": "libavutil56"
			// https://packages.debian.org/bullseye/libavcodec-dev
			"libavcodec-dev": "libavcodec58"
			// https://packages.debian.org/bullseye/libavformat-dev
			"libavformat-dev": "libavformat58"
		}

		build: {
			web: node.#ViteBuild & {
				source: _webappSource.output
				image: mirror: client.env.LINUX_MIRROR
				run: env: GH_PASSWORD: client.env.GH_PASSWORD
				node: npmrc: """
					//npm.pkg.github.com/:_authToken=${GH_PASSWORD}
					@innoai-tech:registry=https://npm.pkg.github.com/
					"""
			}

			golang.#Build & {
				source: _goSource.output
				image: mirror: client.env.LINUX_MIRROR
				image: packages: {
					for _pkgDev, _pkgRun in _deps {
						"\(_pkgDev)": _
					}
				}
				go: {
					cgo:     true
					package: "./cmd/mtk"
					arch:    _archs
					os: ["linux"]
					ldflags: [
						"-s -w",
						"-X \(info.module)/pkg/version.Version=\(_version)",
						"-X \(info.module)/pkg/version.Revision=\(client.env.GIT_SHA)",
					]
				}
				run: {
					env: {
						GOPROXY:   client.env.GOPROXY
						GOPRIVATE: client.env.GOPRIVATE
						GOSUMDB:   client.env.GOSUMDB
					}
					mounts: {
						webui: core.#Mount & {
							dest:     "/go/src/internal/liveplayer/dist"
							contents: web.output
						}
					}
				}
			}
		}

		export: tool.#Export & {
			inputs: {
				for _os in build.go.os for _arch in build.go.arch {
					"\(build.go.name)_\(_os)_\(_arch)": build["\(_os)/\(_arch)"].output
				}
			}
		}

		image: {
			for _arch in _archs {
				"linux/\(_arch)": docker.#Build & {
					steps: [
						debian.#Build & {
							mirror:   client.env.LINUX_MIRROR
							platform: "linux/\(_arch)"
							packages: {
								"ca-certificates": _
								for _pkgDev, _pkgRun in _deps {
									"\(_pkgRun)": _
								}
							}
						},
						docker.#Copy & {
							contents: build["linux/\(_arch)"].output
							source:   "./"
							dest:     "/"
						},
						docker.#Set & {
							config: {
								label: {
									"org.opencontainers.image.source":   "https://\(info.module)"
									"org.opencontainers.image.revision": "\(client.env.GIT_SHA)"
								}
								workdir: "/"
								cmd: ["serve"]
								entrypoint: ["/mtk"]
							}
						},
					]
				}
			}
		}

		push: docker.#Push & {
			dest: "\(_imageName):\(_tag)"
			images: {
				for _arch in _archs {
					"linux/\(_arch)": image["linux/\(_arch)"].output
				}
			}
			auth: {
				username: client.env.GH_USERNAME
				secret:   client.env.GH_PASSWORD
			}
		}
	}
}
