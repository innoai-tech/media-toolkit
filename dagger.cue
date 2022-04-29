package main

import (
	"dagger.io/dagger"
	"dagger.io/dagger/core"
	"universe.dagger.io/docker"

	"github.com/innoai-tech/media-toolkit/cuepkg/tool"
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

		_env: {
			for k, v in client.env if k != "$dagger" {
				"\(k)": v
			}
		}

		_buildPlatform: {
			for k, v in client.platform if k != "$dagger" {
				"\(k)": v
			}
		}

		_imageName: "ghcr.io/innoai-tech/media-toolkit"
		_version:   tool.#ParseVersion & {_, #ref: _env.GIT_REF, #version: _env.VERSION}
		_tag:       _version

		info: tool.#GoModInfo & {
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
			web: tool.#ViteBuild & {
				env: {
					GH_PASSWORD: _env.GH_PASSWORD
				}
				source: _webappSource.output
			}

			tool.#GoBuild & {
				cgoEnabled:    true
				source:        _goSource.output
				package:       "./cmd/mtk"
				buildPlatform: _buildPlatform
				targetPlatform: {
					arch: _archs
					os: ["linux"]
				}
				image: {
					packages: {
						for _pkgDev, _pkgRun in _deps {
							"\(_pkgDev)": _
						}
					}
				}
				run: {
					env: _env
					mounts: {
						webui: core.#Mount & {
							dest:     "/go/src/internal/liveplayer/dist"
							contents: web.output
						}
					}
				}
				ldflags: [
					"-s -w",
					"-X \(info.module)/version.Version=\(_version)",
					"-X \(info.module)/version.Revision=\(_env.GIT_SHA)",
				]
			}
		}

		export: tool.#Export & {
			inputs: {
				for _os in build.targetPlatform.os for _arch in build.targetPlatform.arch {
					"\(build.name)_\(_os)_\(_arch)": build["\(_os)/\(_arch)"].output
				}
			}
		}

		image: {
			for _arch in _archs {
				"linux/\(_arch)": docker.#Build & {
					steps: [
						tool.#DebianBuild & {
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
									"org.opencontainers.image.revision": "\(_env.GIT_SHA)"
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
				username: _env.GH_USERNAME
				secret:   _env.GH_PASSWORD
			}
		}
	}
}
