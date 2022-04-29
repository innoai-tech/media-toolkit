package kube

import (
	core_v1 "k8s.io/api/core/v1"
)

#Namespace: {
	#namespace: string

	core_v1.#Namespace & {
		metadata: name: "\(#namespace)"
	}
}
