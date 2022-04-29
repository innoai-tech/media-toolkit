package mtk

import (
	"encoding/json"

	"github.com/innoai-tech/runtime/cuepkg/kube"
)

#MTK: {
	#values: {
		streams: [ID=string]: {
			id:   ID
			name: string
			rtsp: string
		}
	}

	kube.#App & {
		app: {
			name:    "media-toolkit"
			version: _ | *"main"
		}

		services: "\(app.name)": {
			selector: "app": app.name
			ports:  containers."media-toolkit".ports
			expose: _ | *{
				type: "NodePort"
			}
		}

		containers: "media-toolkit": {
			image: {
				name: _ | *"ghcr.io/innoai-tech/media-toolkit"
				tag:  _ | *"\(app.version)"
			}
			args: ["serve", "--config", "\(volumes.streams.mountPath)", "--addr", ":\(ports.http)"]
			ports: http: _ | *30777
			readinessProbe: kube.#ProbeHttpGet & {
				httpGet: {path: "/", port: ports.http}
			}
			livenessProbe: readinessProbe
		}

		volumes: storage: #MediaDBStorage
		volumes: "streams": {
			mountPath: "/etc/streams.json"
			subPath:   "streams.json"
			source: {
				type: "configMap"
				name: "\(app.name)-config"
				spec: data: {
					"streams.json": json.Marshal([ for _, s in 	#values.streams {s}])
				}
			}
		}
	}
}

#MediaDBStorage: kube.#Volume & {
	mountPath: "/.tmp/mediadb"
	source: {
		claimName: "storage-mediadb"
		spec: {
			accessModes: ["ReadWriteOnce"]
			resources: requests: storage: "10Gi"
			storageClassName: "local-path"
		}
	}
}
