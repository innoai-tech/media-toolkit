package kube

import (
	core_v1 "k8s.io/api/core/v1"
)

#EnvVarSource: {
	configMap: [Name=string]:     string
	secret: [Name=string]:        string
	field: [Name=string]:         string
	resourceField: [Name=string]: string
}

#ProbeHttpGet: core_v1.#Probe & {
	httpGet: {
		scheme: _ | *"HTTP"
	}
	initialDelaySeconds: _ | *5
	timeoutSeconds:      _ | *1
	periodSeconds:       _ | *10
	successThreshold:    _ | *1
	failureThreshold:    _ | *3
}

_containerTemplate: {
	#volumes: [VolumeName=string]: {
		mount:  core_v1.#VolumeMount
		volume: core_v1.#Volume

		for type, v in volume {
			if type == "serect" || type == "configMap" {
				data: [K=string]: string
			}
		}
	}

	#initContainers: [ContainerName=string]: _container & {#name: "\(ContainerName)"}
	#containers: [ContainerName=string]:     _container & {#name: "\(ContainerName)"}

	_container: core_v1.#Container & {
		#name: string
		name:  #name

		#ports: [string]:   int
		#envVars: [string]: string | #EnvVarSource

		ports: [
			for n, cp in #ports {
				name:          n
				containerPort: cp
			},
		]

		env: [
			for _name, _value in #envVars {
				let _isStrValue = (_value & string) != _|_

				[
					if (_isStrValue) {
						name:  _name
						value: _value
					},
					if (!_isStrValue) {
						name: _name
						valueFrom: {
							for _type, _refKey in _value for _ref, _key in _refKey {
								[
									if _type == "secret" {
										secretKeyRef: {
											name: _ref
											key:  [ if _key == "" {_name}, _key][0]
										}
									},
									if _type == "configMap" {
										configMapKeyRef: {
											name: _ref
											key:  [ if _key == "" {_name}, _key][0]
										}
									},
									if _type == "field" {
										fieldRef: {
											fieldPath: _ref
										}
									},
									if _type == "resourceField" {
										resourceFieldRef: {
											resource: _ref
										}
									},
								][0]
							}
						}
					},
				][0]
			},
		]
		volumeMounts: [
			for n, vol in #volumes {
				vol.mount
				name: n
			},
		]
	}

	spec: template: spec: volumes: [
		for n, v in #volumes {
			v.volume
			name: n
		},
	]

	spec: template: spec: initContainers: [
		for c in #initContainers {c},
	]

	spec: template: spec: containers: [
		for c in #containers {c},
	]
}

_derivedManifests: {
	#name:     string
	#workload: _

	// secret or configMap
	for n, v in #workload.spec.template.spec.volumes {
		if v.data != _|_ {
			for volumeSourceName, volumeSource in v.volume {
				if volumeSourceName == "serect" {
					serects: "\(n)": v.data
				}
				if volumeSourceName == "configMap" {
					configMaps: "\(n)": v.data
				}
			}
		}
	}

	// services
	if len(#workload.spec.template.spec.containers) > 0 && len(#workload.spec.template.spec.containers[0].ports) > 0 {
		services: "\(#name)": core_v1.#Service & {
			spec: selector: #workload.spec.selector.matchLabels
			spec: ports: [
				for c in #workload.spec.template.spec.containers
				for p in c.ports {
					name:       *p.name | string
					port:       *p.containerPort | int
					targetPort: *p.containerPort | int
				},
			]
		}
	}
}
