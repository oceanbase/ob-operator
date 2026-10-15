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
{{- if and $service $service.spec.clusterIP -}}
{{- if contains ":" $service.spec.clusterIP -}}
{{- printf "http://[%s]:9000" $service.spec.clusterIP -}}
{{- else -}}
{{- printf "http://%s:9000" $service.spec.clusterIP -}}
{{- end -}}
{{- else -}}{{ include "lakehouse.endpoint" . }}{{- end -}}
{{- else -}}{{ include "lakehouse.endpoint" . }}{{- end -}}
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
{{- define "lakehouse.minioSecret" -}}
{{- default (printf "%s-minio-root" .Release.Name) .Values.minio.rootExistingSecret -}}
{{- end -}}
{{- define "lakehouse.dbSecret" -}}
{{- default (printf "%s-db-root" .Release.Name) .Values.observer.rootExistingSecret -}}
{{- end -}}
{{- define "lakehouse.identity" -}}
{{- dict "mode" .Values.objectStorage.mode "endpoint" (include "lakehouse.endpoint" .) "region" .Values.objectStorage.region "logs" (include "lakehouse.logBucket" .) "data" (include "lakehouse.dataBucket" .) "secret" (include "lakehouse.storageSecret" .) "minioSecret" (include "lakehouse.minioSecret" .) "claim" .Values.minio.existingClaim "size" .Values.minio.size "storageClass" .Values.storageClass "dbSecret" (include "lakehouse.dbSecret" .) "lsId" .Values.logService.clusterId "obId" .Values.observer.clusterId | toJson | sha256sum -}}
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
{{- if eq (include "lakehouse.logBucket" .) (include "lakehouse.dataBucket" .) }}{{ fail "logBucket and dataBucket must differ" }}{{ end -}}
{{- if eq .Values.objectStorage.mode "external" -}}
{{- if not .Values.objectStorage.endpoint }}{{ fail "external storage requires objectStorage.endpoint" }}{{ end -}}
{{- if not .Values.objectStorage.existingSecret }}{{ fail "external storage requires objectStorage.existingSecret" }}{{ end -}}
{{- else -}}
{{- if .Values.objectStorage.existingSecret }}{{ fail "bundled storage generates its own restricted business identity; do not set objectStorage.existingSecret" }}{{ end -}}
{{- if .Values.objectStorage.endpoint }}{{ fail "bundled storage endpoint is generated; do not set objectStorage.endpoint" }}{{ end -}}
{{- end -}}
{{- $guard := lookup "v1" "ConfigMap" .Release.Namespace (printf "%s-storage-identity" .Release.Name) -}}
{{- if $guard -}}
{{- if ne (index $guard.data "identity") (include "lakehouse.identity" .) }}{{ fail "storage identity is immutable: endpoint/buckets/credential names/claim/storage class/cluster IDs must not change; use a new release and an explicit migration" }}{{ end -}}
{{- end -}}
{{- if eq .Values.stage "storage" -}}
{{- if lookup "oceanbase.oceanbase.com/v1alpha1" "OBLogServiceCluster" .Release.Namespace (printf "%s-ls" .Release.Name) }}{{ fail "cannot downgrade stage to storage after LogService creation; use install.sh to resume" }}{{ end -}}
{{- end -}}
{{- if ne .Values.stage "cluster" -}}
{{- if lookup "oceanbase.oceanbase.com/v1alpha1" "OBCluster" .Release.Namespace (printf "%s-ob" .Release.Name) }}{{ fail "cannot downgrade stage after OB creation; use install.sh to resume" }}{{ end -}}
{{- end -}}
{{- end -}}
