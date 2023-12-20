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

.PHONY: deploy-dev
deploy: manifests kustomize ## Deploy dev mode controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/dev | kubectl apply -f -

.PHONY: undeploy-dev
deploy: manifests kustomize ## Undeploy dev mode controller from the K8s cluster specified in ~/.kube/config
	$(KUSTOMIZE) build config/dev | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: redeploy
redeploy: undeploy uninstall export-crd export-operator install deploy ## redeploy crd and controller to the K8s cluster specified in ~/.kube/config.

.PHONY: export-crd
export-crd: manifests kustomize ## Export CRDs manifests
	$(KUSTOMIZE) build config/crd > deploy/crd.yaml

.PHONY: export-operator
export-operator: manifests kustomize ## Export operator manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > deploy/operator.yaml

