{{- define "lakehouse.proxyROSecret" -}}
{{- $cluster := include "lakehouse.lookupResource" (dict "root" . "kind" "OBCluster" "name" (printf "%s-ob" .Release.Name)) | fromJson -}}
{{- dig "spec" "userSecrets" "proxyro" (printf "%s-proxyro" .Release.Name) $cluster -}}
{{- end -}}
