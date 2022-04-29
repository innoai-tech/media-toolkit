package kube

import (
	core_v1 "k8s.io/api/core/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
)

_serviceAccount: {
	#namespace: string

	serviceAccounts: [Name=string]: core_v1.#ServiceAccount & {
		#role: "ClusterRole" | "Role"
		#rules: [...rbac_v1.#PolicyRule]
		metadata: name: Name
	}

	for n, sa in serviceAccounts {
		if sa.#role == "ClusterRole" {
			{
				clusterRoles: "\(n)": rbac_v1.#ClusterRole & {
					metadata: name: "\(n)"
					rules: sa.#rules
				}

				clusterRoleBindings: "\(n)": rbac_v1.#ClusterRoleBinding & {
					metadata: name:      "\(n)"
					metadata: namespace: #namespace

					subjects: [{
						kind:      "ServiceAccount"
						name:      "\(n)"
						namespace: #namespace
					}]

					roleRef: {
						kind:     "ClusterRole"
						name:     "\(n)"
						apiGroup: "rbac.authorization.k8s.io"
					}
				}
			}
		}

		if sa.#role == "Role" {
			{
				roles: "\(n)": rbac_v1.#Role & {
					metadata: name:      "\(n)"
					metadata: namespace: "\(#namespace)"
					rules: sa.#rules
				}

				roleBindings: "\(n)": rbac_v1.#RoleBinding & {
					metadata: name:      "\(n)"
					metadata: namespace: "\(#namespace)"

					subjects: [
						{
							kind:      "ServiceAccount"
							name:      "\(n)"
							namespace: "\(#namespace)"
						},
					]

					roleRef: {
						kind:     "Role"
						name:     "\(n)"
						apiGroup: "rbac.authorization.k8s.io"
					}
				}
			}
		}
	}
}
