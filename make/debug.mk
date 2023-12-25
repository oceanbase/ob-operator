# All targets here are phony
##@ Debug

.PHONY: connect root-pwd

NS ?= default

connect: root-pwd ## Connect to tenant ${TENANT} of cluster ${CLUSTER} with root user
	$(eval nodeHost = $(shell kubectl -n ${NS} get pods -l ref-obcluster=$(clusterName) -o jsonpath='{.items[0].status.podIP}'))
ifdef TENANT
	$(eval secretName = $(shell kubectl -n ${NS} get obtenant ${TENANT} -o jsonpath='{.status.credentials.root}'))
	$(eval tenantName = $(shell kubectl -n ${NS} get obtenant ${TENANT} -o jsonpath='{.spec.tenantName}'))
	$(if $(strip $(secretName)), $(eval pwd = $(shell kubectl -n ${NS} get secret $(secretName) -o jsonpath='{.data.password}' | base64 -d)), )
	$(if $(strip $(pwd)), mysql -h$(nodeHost) -P2881 -A -uroot@$(tenantName) -p$(pwd) -Doceanbase, mysql -h$(nodeHost) -P2881 -A -uroot@$(tenantName) -Doceanbase)
else
	mysql -h$(nodeHost) -P2881 -A -uroot -p$(pwd)
endif

root-pwd: ## Get root password of sys root of cluster ${CLUSTER}
ifdef CLUSTER
	$(eval clusterName = ${CLUSTER})
else
	$(eval clusterName = $(shell kubectl -n ${NS} get obcluster -o jsonpath='{.items[0].metadata.name}'))
endif
	@echo clusterName $(clusterName)
	$(eval secretName = $(shell kubectl -n ${NS} get obcluster $(clusterName) -o jsonpath='{.spec.userSecrets.root}'))
	$(eval nodeHost = $(shell kubectl -n ${NS} get pods -l ref-obcluster=$(clusterName) -o jsonpath='{.items[0].status.podIP}'))
	$(if $(strip $(secretName)), $(eval pwd = $(shell kubectl -n ${NS} get secret $(secretName) -o jsonpath='{.data.password}' | base64 -d)), )
	@echo root pwd of sys tenant of cluster '$(clusterName)' is $(pwd)

