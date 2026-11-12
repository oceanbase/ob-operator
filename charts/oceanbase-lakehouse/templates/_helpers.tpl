{{- define "lakehouse.storageMode" -}}
{{- if .Values.objectStorage.endpoint -}}external{{- else -}}bundled{{- end -}}
{{- end -}}
{{- define "lakehouse.name" -}}
{{- .Release.Name -}}
{{- end -}}
{{- define "lakehouse.labels" -}}
app.kubernetes.io/name: oceanbase-lakehouse
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}
{{- define "lakehouse.scheduling" -}}
{{- with .Values.scheduling.nodeSelector }}
nodeSelector: {{ toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.scheduling.tolerations }}
tolerations: {{ toYaml . | nindent 2 }}
{{- end }}
{{- end -}}

{{/* Zone selectors override defaults without mutating other zones or storage. */}}
{{- define "lakehouse.zoneScheduling" -}}
{{- $selector := mergeOverwrite (deepCopy .root.Values.scheduling.nodeSelector) (.zone.nodeSelector | default dict) -}}
{{- with $selector }}
nodeSelector: {{ toYaml . | nindent 2 }}
{{- end }}
{{- with .root.Values.scheduling.tolerations }}
tolerations: {{ toYaml . | nindent 2 }}
{{- end }}
{{- end -}}

{{- define "lakehouse.endpoint" -}}
{{- if eq (include "lakehouse.storageMode" .) "bundled" -}}
{{- printf "http://%s-minio.%s.svc:9000" .Release.Name .Release.Namespace -}}
{{- else -}}{{ .Values.objectStorage.endpoint }}{{- end -}}
{{- end -}}
{{- define "lakehouse.databaseEndpoint" -}}
{{- if eq (include "lakehouse.storageMode" .) "bundled" -}}
{{- $service := lookup "v1" "Service" .Release.Namespace (printf "%s-minio" .Release.Name) -}}
{{- if .Values.native.enabled -}}
{{- include "lakehouse.endpoint" . -}}
{{- else if and $service $service.spec.clusterIP -}}
{{- if contains ":" $service.spec.clusterIP -}}
{{- printf "http://[%s]:9000" $service.spec.clusterIP -}}
{{- else -}}
{{- printf "http://%s:9000" $service.spec.clusterIP -}}
{{- end -}}
{{- else -}}{{ include "lakehouse.endpoint" . }}{{- end -}}
{{- else -}}{{ include "lakehouse.externalDatabaseEndpoint" . }}{{- end -}}
{{- end -}}
{{/* LS 1.3.0 ignores path_style for DNS hosts. Resolve only a same-cluster
HTTP Service, never public DNS or HTTPS (which would change TLS identity).
Old releases keep their exact URL; new ones pin the resolved address. */}}
{{- define "lakehouse.externalDatabaseEndpoint" -}}
{{- $endpoint := .Values.objectStorage.endpoint -}}
{{- $guard := lookup "v1" "ConfigMap" .Release.Namespace (printf "%s-storage-identity" .Release.Name) -}}
{{- if and $guard (not (hasKey $guard.data "externalDatabaseEndpoint")) -}}
{{- $endpoint -}}
{{- else -}}
{{- $url := urlParse $endpoint -}}
{{- $host := regexReplaceAll ":[0-9]+$" $url.host "" -}}
{{- if regexMatch "^[a-z0-9-]+\\.[a-z0-9-]+\\.svc(\\.cluster\\.local)?$" $host -}}
{{- if ne $url.scheme "http" }}{{ fail "LogService 1.3.0 cannot use this HTTPS Kubernetes Service DNS endpoint without bucket DNS; refusing to replace TLS hostname with an IP" }}{{ end -}}
{{- $parts := splitList "." $host -}}
{{- $service := lookup "v1" "Service" (index $parts 1) (index $parts 0) -}}
{{- if not $service }}{{ fail "external Kubernetes Service not found; verify endpoint and Helm Service read permissions" }}{{ end -}}
{{- $ip := $service.spec.clusterIP | default "" -}}
{{- if or (eq $ip "") (eq $ip "None") }}{{ fail "external storage requires a stable ClusterIP Service; headless and ExternalName Services are unsupported" }}{{ end -}}
{{- $port := 80 -}}
{{- if contains ":" $url.host }}{{ $port = int (last (splitList ":" $url.host)) }}{{ end -}}
{{- $found := false -}}
{{- range $service.spec.ports -}}
{{- if and (eq (int .port) $port) (eq (.protocol | default "TCP") "TCP") }}{{ $found = true }}{{ end -}}
{{- end -}}
{{- if not $found }}{{ fail "external storage endpoint port is not exposed by the Kubernetes Service" }}{{ end -}}
{{- if contains ":" $ip }}{{ $ip = printf "[%s]" $ip }}{{ end -}}
{{- $endpoint = printf "http://%s:%d" $ip $port -}}
{{- end -}}
{{- if and $guard (ne (index $guard.data "externalDatabaseEndpoint") $endpoint) }}{{ fail "external storage Service address changed; refusing to redirect retained database data; restore the original Service or perform an explicit migration" }}{{ end -}}
{{- $endpoint -}}
{{- end -}}
{{- end -}}
{{- define "lakehouse.preflightEndpoint" -}}
{{- if eq (include "lakehouse.storageMode" .) "external" }}{{ include "lakehouse.databaseEndpoint" . }}{{ else }}{{ include "lakehouse.endpoint" . }}{{ end -}}
{{- end -}}
{{- define "lakehouse.preflightPath" -}}
{{- $host := (urlParse (include "lakehouse.preflightEndpoint" .)).host -}}
{{- if or (eq (include "lakehouse.storageMode" .) "bundled") (regexMatch "^([0-9]+\\.){3}[0-9]+(:[0-9]+)?$" $host) (hasPrefix "[" $host) -}}on{{- else -}}off{{- end -}}
{{- end -}}
{{- define "lakehouse.logBucket" -}}
{{- default (printf "%s-logs" .Release.Name) .Values.objectStorage.logBucket -}}
{{- end -}}
{{- define "lakehouse.dataBucket" -}}
{{- default (printf "%s-data" .Release.Name) .Values.objectStorage.dataBucket -}}
{{- end -}}
{{- define "lakehouse.storageSecret" -}}
{{- default (printf "%s-storage" .Release.Name) .Values.objectStorage.existingSecret -}}
{{- end -}}
{{- define "lakehouse.logStorageSecret" -}}
{{- if eq (include "lakehouse.storageMode" .) "bundled" -}}
{{- include "lakehouse.storageSecret" . -}}
{{- else -}}{{ default .Values.objectStorage.existingSecret .Values.objectStorage.logExistingSecret }}{{- end -}}
{{- end -}}
{{- define "lakehouse.dataStorageSecret" -}}
{{- if eq (include "lakehouse.storageMode" .) "bundled" -}}
{{- include "lakehouse.storageSecret" . -}}
{{- else -}}{{ default .Values.objectStorage.existingSecret .Values.objectStorage.dataExistingSecret }}{{- end -}}
{{- end -}}
{{- define "lakehouse.minioSecret" -}}
{{- default (printf "%s-minio-root" .Release.Name) .Values.minio.rootExistingSecret -}}
{{- end -}}
{{- define "lakehouse.dbSecret" -}}
{{- default (printf "%s-db-root" .Release.Name) .Values.observer.rootExistingSecret -}}
{{- end -}}
{{- define "lakehouse.databaseVolumeClasses" -}}
{{- dict
  "observerData" (default .Values.storageClass .Values.observer.dataStorageClass)
  "observerLog" (default .Values.storageClass .Values.observer.logStorageClass)
  "logServiceStore" (default .Values.storageClass .Values.logService.storeStorageClass)
  "logServiceLog" (default .Values.storageClass .Values.logService.logStorageClass)
  | toJson -}}
{{- end -}}
{{- define "lakehouse.extendStorageIdentity" -}}
{{- $classes := include "lakehouse.databaseVolumeClasses" .root | fromJson -}}
{{- range $name, $class := $classes -}}
  {{- if ne $class $.root.Values.storageClass -}}
    {{- $_ := set $.identity $name $class -}}
  {{- end -}}
{{- end -}}
{{- if and (eq (include "lakehouse.storageMode" .root) "bundled") (eq .root.Values.minio.mode "distributed") -}}
  {{- $_ := set .identity "minioDistributed" .root.Values.minio.distributed -}}
{{- end -}}
{{- end -}}
{{- define "lakehouse.identity" -}}
{{- $identity := dict "mode" (include "lakehouse.storageMode" .) "endpoint" (include "lakehouse.endpoint" .) "region" .Values.objectStorage.region "logs" (include "lakehouse.logBucket" .) "data" (include "lakehouse.dataBucket" .) "secret" (include "lakehouse.storageSecret" .) "minioSecret" (include "lakehouse.minioSecret" .) "claim" .Values.minio.existingClaim "size" .Values.minio.size "storageClass" .Values.storageClass "dbSecret" (include "lakehouse.dbSecret" .) "lsId" .Values.logService.clusterId "obId" .Values.observer.clusterId -}}
{{/* Keep the exact legacy hash when both effective references are unchanged.
Actual per-side overrides are immutable, including after uninstall/reinstall. */}}
{{- if or (ne (include "lakehouse.logStorageSecret" .) (include "lakehouse.storageSecret" .)) (ne (include "lakehouse.dataStorageSecret" .) (include "lakehouse.storageSecret" .)) -}}
{{- $_ := set $identity "logSecret" (include "lakehouse.logStorageSecret" .) -}}
{{- $_ := set $identity "dataSecret" (include "lakehouse.dataStorageSecret" .) -}}
{{- end -}}
{{- $guard := lookup "v1" "ConfigMap" .Release.Namespace (printf "%s-storage-identity" .Release.Name) -}}
{{- if and (eq (include "lakehouse.storageMode" .) "external") (or (not $guard) (hasKey $guard.data "externalDatabaseEndpoint")) -}}
{{- $_ := set $identity "externalDatabaseEndpoint" (include "lakehouse.databaseEndpoint" .) -}}
{{- end -}}
{{- include "lakehouse.extendStorageIdentity" (dict "root" . "identity" $identity) -}}
{{- $identity | toJson | sha256sum -}}
{{- end -}}
{{- define "lakehouse.databaseIdentity" -}}
{{- $identity := dict "mode" (include "lakehouse.storageMode" .) "endpoint" (include "lakehouse.endpoint" .) "region" .Values.objectStorage.region "logs" (include "lakehouse.logBucket" .) "data" (include "lakehouse.dataBucket" .) "logSecret" (include "lakehouse.logStorageSecret" .) "dataSecret" (include "lakehouse.dataStorageSecret" .) "storageClass" .Values.storageClass "dbSecret" (include "lakehouse.dbSecret" .) "lsId" .Values.logService.clusterId "obId" .Values.observer.clusterId -}}
{{- if eq (include "lakehouse.storageMode" .) "bundled" -}}
{{- $_ := set $identity "minio" (dict "secret" (include "lakehouse.minioSecret" .) "claim" .Values.minio.existingClaim "size" .Values.minio.size "storageClass" .Values.minio.storageClass) -}}
{{- end -}}
{{- if eq (include "lakehouse.storageMode" .) "external" -}}
{{- $_ := set $identity "externalDatabaseEndpoint" (include "lakehouse.databaseEndpoint" .) -}}
{{- end -}}
{{- include "lakehouse.extendStorageIdentity" (dict "root" . "identity" $identity) -}}
{{- $identity | toJson | sha256sum -}}
{{- end -}}
{{- define "lakehouse.validate" -}}
{{- if and (not .Values.objectStorage.endpoint) (not .Values.minio.enabled) }}{{ fail "configure objectStorage.endpoint or enable minio.enabled" }}{{ end -}}
{{- if and (not .Values.objectStorage.endpoint) (ne .Values.objectStorage.region .Values.minio.region) }}{{ fail "internal storage requires matching objectStorage.region and minio.region; for legacy upgrades run tools/migrate-minio-values.py on helm get values --all" }}{{ end -}}
{{- if gt (len .Release.Name) 40 }}{{ fail "release name must be at most 40 characters" }}{{ end -}}
{{- range $topology := list .Values.logService.topology .Values.observer.topology -}}
{{- $zones := dict -}}
{{- range $zone := $topology -}}
{{- if hasKey $zones $zone.zone }}{{ fail "topology contains duplicate zones" }}{{ end -}}
{{- $_ := set $zones $zone.zone true -}}
{{- end -}}
{{- end -}}
{{- $existingLS := (include "lakehouse.lookupResource" (dict "root" . "kind" "OBLogServiceCluster" "name" (printf "%s-ls" .Release.Name)) | fromJson) -}}
{{- $existingOB := (include "lakehouse.lookupResource" (dict "root" . "kind" "OBCluster" "name" (printf "%s-ob" .Release.Name)) | fromJson) -}}
{{- $classes := include "lakehouse.databaseVolumeClasses" . | fromJson -}}
{{- $volumes := list
  (dict "resource" $existingOB "component" "observer" "volume" "dataStorage" "expected" $classes.observerData)
  (dict "resource" $existingOB "component" "observer" "volume" "logStorage" "expected" $classes.observerLog)
  (dict "resource" $existingLS "component" "logService" "volume" "storeStorage" "expected" $classes.logServiceStore)
  (dict "resource" $existingLS "component" "logService" "volume" "logStorage" "expected" $classes.logServiceLog) -}}
{{- range $volumes -}}
  {{- $actual := dig "spec" .component "storage" .volume "storageClass" "" .resource -}}
  {{- if and $actual (ne $actual .expected) -}}
    {{- fail (printf "%s.%s StorageClass is immutable for existing storage; use an explicit data migration" .component .volume) -}}
  {{- end -}}
{{- end -}}
{{- if not $existingLS -}}
{{- $nodes := 0 -}}
{{- range .Values.logService.topology }}{{ $nodes = add $nodes (int .replica) }}{{ end -}}
{{- if ne (int $nodes) 3 }}{{ fail "LogService 1.3.0 bootstrap requires exactly 3 nodes; configure topology replicas totaling 3 for a new installation" }}{{ end -}}
{{- end -}}
{{- if and (eq (include "lakehouse.storageMode" .) "bundled") (eq (include "lakehouse.logBucket" .) (include "lakehouse.dataBucket" .)) }}{{ fail "logBucket and dataBucket must differ in bundled mode" }}{{ end -}}
{{- if eq (include "lakehouse.storageMode" .) "external" -}}
{{- if not .Values.objectStorage.endpoint }}{{ fail "external storage requires objectStorage.endpoint" }}{{ end -}}
{{- if not (include "lakehouse.logStorageSecret" .) }}{{ fail "external storage requires objectStorage.logExistingSecret or objectStorage.existingSecret" }}{{ end -}}
{{- if not (include "lakehouse.dataStorageSecret" .) }}{{ fail "external storage requires objectStorage.dataExistingSecret or objectStorage.existingSecret" }}{{ end -}}
{{- else -}}
{{- if or .Values.objectStorage.logExistingSecret .Values.objectStorage.dataExistingSecret }}{{ fail "per-side credential references require external storage mode" }}{{ end -}}
{{- if .Values.objectStorage.existingSecret }}{{ fail "bundled storage generates its own restricted business identity; do not set objectStorage.existingSecret" }}{{ end -}}
{{- if .Values.objectStorage.endpoint }}{{ fail "bundled storage endpoint is generated; do not set objectStorage.endpoint" }}{{ end -}}
{{- end -}}
{{- $guard := lookup "v1" "ConfigMap" .Release.Namespace (printf "%s-storage-identity" .Release.Name) -}}
{{- if $guard -}}
{{- $expected := include "lakehouse.identity" . -}}
{{- $actual := index $guard.data "identity" -}}
{{- if hasKey $guard.data "databaseIdentity" -}}
{{- $expected = include "lakehouse.databaseIdentity" . -}}
{{- $actual = index $guard.data "databaseIdentity" -}}
{{- end -}}
{{- if ne $actual $expected }}{{ fail "storage identity is immutable: endpoint/buckets/credential names/claim/storage class/cluster IDs must not change; use a new release and an explicit migration" }}{{ end -}}
{{- end -}}
{{/* Before adopting old resources, require explicit preservation of the parent
settings which the former MinIO template inherited. No template mutates values. */}}
{{- if and $guard (not (hasKey $guard.data "databaseIdentity")) (eq (include "lakehouse.storageMode" .) "bundled") -}}
{{- $different := and (not .Values.minio.existingClaim) (ne .Values.minio.storageClass .Values.storageClass) -}}
{{- range $field := list "nodeSelector" "tolerations" -}}
{{- if ne (toJson (index $.Values.minio $field)) (toJson (index $.Values.scheduling $field)) -}}{{- $different = true -}}{{- end -}}
{{- end -}}
{{- if ne (toJson .Values.minio.imagePullSecrets) (toJson .Values.imagePullSecrets) -}}{{- $different = true -}}{{- end -}}
{{- if $different }}{{ fail "legacy MinIO configuration requires migration before upgrade: export helm get values --all and run tools/migrate-minio-values.py to preserve storageClass, region, scheduling and imagePullSecrets" }}{{ end -}}
{{- end -}}
{{- if eq .Values.stage "storage" -}}
{{- if (include "lakehouse.lookupResource" (dict "root" . "kind" "OBLogServiceCluster" "name" (printf "%s-ls" .Release.Name)) | fromJson) }}{{ fail "cannot downgrade stage to storage after LogService creation; use helm upgrade with stage=cluster to resume" }}{{ end -}}
{{- end -}}
{{- if ne .Values.stage "cluster" -}}
{{- if (include "lakehouse.lookupResource" (dict "root" . "kind" "OBCluster" "name" (printf "%s-ob" .Release.Name)) | fromJson) }}{{ fail "cannot downgrade stage after OB creation; use helm upgrade with stage=cluster to resume" }}{{ end -}}
{{- end -}}
{{- end -}}

{{/* Preserve immutable CR URLs when adopting a release created by the legacy
staged installer. New native releases keep DNS in the parent CR, and Operator
resolves its same-namespace Service at runtime without mutating this manifest. */}}
{{- define "lakehouse.logBucketURL" -}}
{{- $existing := (include "lakehouse.lookupResource" (dict "root" . "kind" "OBLogServiceCluster" "name" (printf "%s-ls" .Release.Name)) | fromJson) -}}
{{- if $existing -}}{{ $existing.spec.objectStoreUrl.bucketURL }}{{- else -}}
{{- printf "s3://%s?host=%s&s3_region=%s" (include "lakehouse.logBucket" .) (include "lakehouse.databaseEndpoint" .) .Values.objectStorage.region -}}
{{- end -}}
{{- end -}}
{{- define "lakehouse.dataBucketURL" -}}
{{- $existing := (include "lakehouse.lookupResource" (dict "root" . "kind" "OBCluster" "name" (printf "%s-ob" .Release.Name)) | fromJson) -}}
{{- if $existing -}}{{ $existing.spec.sharedStorageInfo.bucketURL }}{{- else -}}
{{- printf "s3://%s?host=%s&s3_region=%s" (include "lakehouse.dataBucket" .) (include "lakehouse.databaseEndpoint" .) .Values.objectStorage.region -}}
{{- end -}}
{{- end -}}

{{- define "lakehouse.lookupResource" -}}
{{- if .root.Capabilities.APIVersions.Has (printf "oceanbase.oceanbase.com/v1alpha1/%s" .kind) -}}
{{- lookup "oceanbase.oceanbase.com/v1alpha1" .kind .root.Release.Namespace .name | toJson -}}
{{- else -}}{}{{- end -}}
{{- end -}}

{{/* Presentation address only: never put a discovered node IP into Service specs. */}}
{{- define "lakehouse.accessHost" -}}
{{- if .Values.access.host -}}
{{- .Values.access.host -}}
{{- else -}}
{{- $host := "" -}}
{{- range (lookup "v1" "Node" "" "").items -}}
  {{- $ready := false -}}
  {{- range .status.conditions -}}
    {{- if and (eq .type "Ready") (eq .status "True") -}}{{- $ready = true -}}{{- end -}}
  {{- end -}}
  {{- if and $ready (not $host) -}}
    {{- $internal := "" -}}{{- $external := "" -}}
    {{- range .status.addresses -}}
      {{- if eq .type "InternalIP" -}}{{- $internal = .address -}}{{- end -}}
      {{- if eq .type "ExternalIP" -}}{{- $external = .address -}}{{- end -}}
    {{- end -}}
    {{- $host = default $external $internal -}}
  {{- end -}}
{{- end -}}
{{- $host -}}
{{- end -}}
{{- end -}}
{{- define "lakehouse.urlHost" -}}
{{- if contains ":" . -}}[{{ . }}]{{- else -}}{{ . }}{{- end -}}
{{- end -}}

{{/* Do not query custom APIs before CRDs exist on first installation. */}}

{{/* Zero allocates a new port on install and preserves this release's port on upgrade. */}}
{{- define "lakehouse.sqlNodePort" -}}
{{- if gt (int .Values.sqlService.nodePort) 0 -}}{{ .Values.sqlService.nodePort }}
{{- else -}}
{{- $service := lookup "v1" "Service" .Release.Namespace (printf "%s-sql" .Release.Name) -}}
{{- if and $service $service.spec.ports -}}{{ (index $service.spec.ports 0).nodePort | default 0 }}{{- else -}}0{{- end -}}
{{- end -}}
{{- end -}}

{{/* Presentation only: prefer an explicit port, otherwise read the existing Service. */}}
{{- define "lakehouse.dashboardNodePort" -}}
{{- if gt (int .Values.dashboard.service.nodePort) 0 -}}{{ .Values.dashboard.service.nodePort }}
{{- else -}}
{{- $service := lookup "v1" "Service" .Release.Namespace (printf "oceanbase-dashboard-%s" .Release.Name) -}}
{{- if and $service $service.spec.ports -}}{{ (index $service.spec.ports 0).nodePort | default 0 }}{{- else -}}0{{- end -}}
{{- end -}}
{{- end -}}
