##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: redeploy
redeploy: undeploy uninstall export-crd export-operator install deploy ## redeploy crd and controller to the K8s cluster specified in ~/.kube/config.

.PHONY: export-operator
export-operator: manifests kustomize ## Export operator manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > deploy/operator.yaml

.PHONY: export-charts
export-charts: export-operator ## Export ob-operator helm chart
	$(KUSTOMIZE) build config/helm | sed \
		-e '1,/---/d' \
		-e 's/oceanbase-system/{{ .Release.Namespace }}/g' \
		-e 's/value: ob-operator/value: {{ .Values.reporter }}/g' \
		-e 's#image: HELM_OPERATOR_IMAGE$$#image: {{ .Values.image | quote }}#' \
		-e 's#image: HELM_KUBE_RBAC_PROXY_IMAGE$$#image: {{ .Values.kubeRbacProxyImage | default "oceanbase/kube-rbac-proxy:v0.13.1" | quote }}#' \
		> charts/ob-operator/templates/operator.yaml
