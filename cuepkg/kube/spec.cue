package kube

import (
	core_v1 "k8s.io/api/core/v1"
	networking_v1 "k8s.io/api/networking/v1"
	apps_v1 "k8s.io/api/apps/v1"
)

#Spec: {
	#namespace: string | *"default"

	services: [Name = string]: core_v1.#Service & {
		metadata: name:      Name
		metadata: namespace: #namespace
	}

	ingresses: [Name = string]: networking_v1.#Ingress & {
		metadata: name:      Name
		metadata: namespace: #namespace
	}

	persistentVolumeClaims: [Name = string]: core_v1.#PersistentVolumeClaim & {
		metadata: name:      Name
		metadata: namespace: #namespace
	}

	configMaps: [Name = string]: core_v1.#ConfigMap & {
		metadata: name:      Name
		metadata: namespace: #namespace
	}

	secrets: [Name = string]: core_v1.#Secret & {
		metadata: name:      Name
		metadata: namespace: #namespace
		type: string | *"Opaque"
	}

	deployments: [Name = string]: apps_v1.#Deployment & {
		metadata: name:      Name
		metadata: namespace: #namespace
		metadata: labels: app: Name
		spec: template: metadata: labels: app: Name
		spec: selector: matchLabels: app: Name
		spec: replicas: int | *1

		_containerTemplate
	}

	daemonSets: [Name = string]: apps_v1.#DaemonSet & {
		metadata: name:      Name
		metadata: namespace: #namespace
		metadata: labels: app: Name
		spec: template: metadata: labels: app: Name
		spec: selector: matchLabels: app: Name

		_containerTemplate
	}

	statefulSets: [Name = string]: apps_v1.#StatefulSet & {
		metadata: name:      Name
		metadata: namespace: #namespace
		metadata: labels: app: Name
		spec: template: metadata: labels: app: Name
		spec: selector: matchLabels: app: Name

		_containerTemplate
	}

	for n, w in deployments {
		_derivedManifests & {#name: n, #workload: w}
	}
	for n, w in statefulSets {
		_derivedManifests & {#name: n, #workload: w}
	}
	for n, w in daemonSets {
		_derivedManifests & {#name: n, #workload: w}
	}

	_namespace: #namespace
	_serviceAccount & {#namespace: _namespace}
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
