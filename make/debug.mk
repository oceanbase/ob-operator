# All targets here are phony
##@ Debug

.PHONY: connect gettenants getpolicy getbackupjobs getcluster getobserver getrestorejobs getpods

connect:
	$(eval nodeHost = $(shell kubectl get pods -o jsonpath='{.items[0].status.podIP}'))
ifdef TENANT
	$(eval secretName = $(shell kubectl get obtenant ${TENANT} -o jsonpath='{.status.credentials.root}'))
	$(eval tenantName = $(shell kubectl get obtenant ${TENANT} -o jsonpath='{.spec.tenantName}'))
	$(if $(strip $(secretName)), $(eval pwd = $(shell kubectl get secret $(secretName) -o jsonpath='{.data.password}' | base64 -d)), )
	$(if $(strip $(pwd)), mysql -h$(nodeHost) -P2881 -A -uroot@$(tenantName) -p$(pwd) -Doceanbase, mysql -h$(nodeHost) -P2881 -A -uroot@$(tenantName) -Doceanbase)
else
	mysql -h$(nodeHost) -P2881 -A -uroot -p -Doceanbase
endif

gettenants: ## Get all tenants
	@kubectl get -n oceanbase obtenant -o wide

getpolicy: ## Get all backup policy
	@kubectl get -n oceanbase obtenantbackuppolicy
	
getbackupjobs: ## Get all backup jobs
	@kubectl get -n oceanbase obtenantbackup

getcluster: ## Get all obclusters
	@kubectl get -n oceanbase obcluster

getobserver: ## Get all observers
	@kubectl get -n oceanbase observer

getrestorejobs:
	@kubectl get -n oceanbase obtenantrestore

getpods:
	@kubectl get -n oceanbase pods -o wide