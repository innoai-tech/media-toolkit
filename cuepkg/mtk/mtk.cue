package kubepkg

import (
	"encoding/json"
)

#Mtk: kube.#Spec & {
	#values: {
		image: {
			name:       string | *"ghcr.io/innoai-tech/media-toolkit"
			tag:        string | *"main"
			pullPolicy: string | *"IfNotPresent"
		}

		http: {
			port: 80 | *int
		}

		streams: [ID=string]: {name: string, rtsp: string}
	}

	#name: "media-toolkit"

	configMaps: "\(#name)-config": data: {
		"streams.json": json.Marshal([
				for id, s in streams {
				{
					s
					id: id
				}
			},
		])
	}

	deployments: "\(#name)": {
		#volumes: {
			storage: {
				mount: mountPath: "/.tmp/mediadb"
				volume: persistentVolumeClaim: claimName: _pvcStorage.#pvcName
			}
			config: {
				mount: mountPath: "/etc/streams.json"
				volume: configMap: name: "\(#name)-config"
			}
		}

		#containers: "media-toolkit": {
			image:           "\(#values.image.name):\(#values.image.tag)"
			imagePullPolicy: "\(#values.image.pullPolicy)"
			args: ["serve", "--config", "/etc/streams.json", "--addr", ":\(#values.http.port)"]

			#ports: {
				http: #values.http.port
			}

			readinessProbe: kube.#ProbeHttpGet & {
				httpGet: {path: "/", port: #ports.http}
			}

			livenessProbe: readinessProbe
		}
	}
}

_pvcStorage: {
	#pvcName: "storage-mtk"

	persistentVolumeClaims: "\(#pvcName)": {
		spec: {
			accessModes: ["ReadWriteOnce"]
			resources: requests: storage: "10Gi"
			storageClassName: "local-path"
		}
	}
}
