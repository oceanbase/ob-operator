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
{{- if eq .Values.objectStorage.mode "bundled" -}}
{{- printf "http://%s-minio.%s.svc:9000" .Release.Name .Release.Namespace -}}
{{- else -}}{{ .Values.objectStorage.endpoint }}{{- end -}}
{{- end -}}
{{- define "lakehouse.databaseEndpoint" -}}
{{- if eq .Values.objectStorage.mode "bundled" -}}
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
{{- if eq .Values.objectStorage.mode "external" }}{{ include "lakehouse.databaseEndpoint" . }}{{ else }}{{ include "lakehouse.endpoint" . }}{{ end -}}
{{- end -}}
{{- define "lakehouse.preflightPath" -}}
{{- $host := (urlParse (include "lakehouse.preflightEndpoint" .)).host -}}
{{- if or (eq .Values.objectStorage.mode "bundled") (regexMatch "^([0-9]+\\.){3}[0-9]+(:[0-9]+)?$" $host) (hasPrefix "[" $host) -}}on{{- else -}}off{{- end -}}
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
{{- if eq .Values.objectStorage.mode "bundled" -}}
{{- include "lakehouse.storageSecret" . -}}
{{- else -}}{{ default .Values.objectStorage.existingSecret .Values.objectStorage.logExistingSecret }}{{- end -}}
{{- end -}}
{{- define "lakehouse.dataStorageSecret" -}}
{{- if eq .Values.objectStorage.mode "bundled" -}}
{{- include "lakehouse.storageSecret" . -}}
{{- else -}}{{ default .Values.objectStorage.existingSecret .Values.objectStorage.dataExistingSecret }}{{- end -}}
{{- end -}}
{{- define "lakehouse.minioSecret" -}}
{{- default (printf "%s-minio-root" .Release.Name) .Values.minio.rootExistingSecret -}}
{{- end -}}
{{- define "lakehouse.dbSecret" -}}
{{- default (printf "%s-db-root" .Release.Name) .Values.observer.rootExistingSecret -}}
{{- end -}}
{{- define "lakehouse.identity" -}}
{{- $identity := dict "mode" .Values.objectStorage.mode "endpoint" (include "lakehouse.endpoint" .) "region" .Values.objectStorage.region "logs" (include "lakehouse.logBucket" .) "data" (include "lakehouse.dataBucket" .) "secret" (include "lakehouse.storageSecret" .) "minioSecret" (include "lakehouse.minioSecret" .) "claim" .Values.minio.existingClaim "size" .Values.minio.size "storageClass" .Values.storageClass "dbSecret" (include "lakehouse.dbSecret" .) "lsId" .Values.logService.clusterId "obId" .Values.observer.clusterId -}}
{{/* Keep the exact legacy hash when both effective references are unchanged.
Actual per-side overrides are immutable, including after uninstall/reinstall. */}}
{{- if or (ne (include "lakehouse.logStorageSecret" .) (include "lakehouse.storageSecret" .)) (ne (include "lakehouse.dataStorageSecret" .) (include "lakehouse.storageSecret" .)) -}}
{{- $_ := set $identity "logSecret" (include "lakehouse.logStorageSecret" .) -}}
{{- $_ := set $identity "dataSecret" (include "lakehouse.dataStorageSecret" .) -}}
{{- end -}}
{{- $guard := lookup "v1" "ConfigMap" .Release.Namespace (printf "%s-storage-identity" .Release.Name) -}}
{{- if and (eq .Values.objectStorage.mode "external") (or (not $guard) (hasKey $guard.data "externalDatabaseEndpoint")) -}}
{{- $_ := set $identity "externalDatabaseEndpoint" (include "lakehouse.databaseEndpoint" .) -}}
{{- end -}}
{{- $identity | toJson | sha256sum -}}
{{- end -}}
{{- define "lakehouse.validate" -}}
{{- if gt (len .Release.Name) 40 }}{{ fail "release name must be at most 40 characters" }}{{ end -}}
{{- range $topology := list .Values.logService.topology .Values.observer.topology -}}
{{- $zones := dict -}}
{{- range $zone := $topology -}}
{{- if hasKey $zones $zone.zone }}{{ fail "topology contains duplicate zones" }}{{ end -}}
{{- $_ := set $zones $zone.zone true -}}
{{- end -}}
{{- end -}}
{{- $existingLS := lookup "oceanbase.oceanbase.com/v1alpha1" "OBLogServiceCluster" .Release.Namespace (printf "%s-ls" .Release.Name) -}}
{{- if not $existingLS -}}
{{- $nodes := 0 -}}
{{- range .Values.logService.topology }}{{ $nodes = add $nodes (int .replica) }}{{ end -}}
{{- if ne (int $nodes) 3 }}{{ fail "LogService 1.3.0 bootstrap requires exactly 3 nodes; configure topology replicas totaling 3 for a new installation" }}{{ end -}}
{{- end -}}
{{- if and (eq .Values.objectStorage.mode "bundled") (eq (include "lakehouse.logBucket" .) (include "lakehouse.dataBucket" .)) }}{{ fail "logBucket and dataBucket must differ in bundled mode" }}{{ end -}}
{{- if eq .Values.objectStorage.mode "external" -}}
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
{{- if ne (index $guard.data "identity") (include "lakehouse.identity" .) }}{{ fail "storage identity is immutable: endpoint/buckets/credential names/claim/storage class/cluster IDs must not change; use a new release and an explicit migration" }}{{ end -}}
{{- end -}}
{{- if eq .Values.stage "storage" -}}
{{- if lookup "oceanbase.oceanbase.com/v1alpha1" "OBLogServiceCluster" .Release.Namespace (printf "%s-ls" .Release.Name) }}{{ fail "cannot downgrade stage to storage after LogService creation; use helm upgrade with stage=cluster to resume" }}{{ end -}}
{{- end -}}
{{- if ne .Values.stage "cluster" -}}
{{- if lookup "oceanbase.oceanbase.com/v1alpha1" "OBCluster" .Release.Namespace (printf "%s-ob" .Release.Name) }}{{ fail "cannot downgrade stage after OB creation; use helm upgrade with stage=cluster to resume" }}{{ end -}}
{{- end -}}
{{- end -}}

{{/* Preserve immutable CR URLs when adopting a release created by the legacy
staged installer. New native releases keep DNS in the parent CR, and Operator
resolves its same-namespace Service at runtime without mutating this manifest. */}}
{{- define "lakehouse.logBucketURL" -}}
{{- $existing := lookup "oceanbase.oceanbase.com/v1alpha1" "OBLogServiceCluster" .Release.Namespace (printf "%s-ls" .Release.Name) -}}
{{- if $existing -}}{{ $existing.spec.objectStoreUrl.bucketURL }}{{- else -}}
{{- printf "s3://%s?host=%s&s3_region=%s" (include "lakehouse.logBucket" .) (include "lakehouse.databaseEndpoint" .) .Values.objectStorage.region -}}
{{- end -}}
{{- end -}}
{{- define "lakehouse.dataBucketURL" -}}
{{- $existing := lookup "oceanbase.oceanbase.com/v1alpha1" "OBCluster" .Release.Namespace (printf "%s-ob" .Release.Name) -}}
{{- if $existing -}}{{ $existing.spec.sharedStorageInfo.bucketURL }}{{- else -}}
{{- printf "s3://%s?host=%s&s3_region=%s" (include "lakehouse.dataBucket" .) (include "lakehouse.databaseEndpoint" .) .Values.objectStorage.region -}}
{{- end -}}
{{- end -}}
