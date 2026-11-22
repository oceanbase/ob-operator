{{/*
Expand the name of the chart.
*/}}
{{- define "oceanbase-dashboard.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "oceanbase-dashboard.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "oceanbase-dashboard.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "oceanbase-dashboard.labels" -}}
helm.sh/chart: {{ include "oceanbase-dashboard.chart" . }}
{{ include "oceanbase-dashboard.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "oceanbase-dashboard.selectorLabels" -}}
app.kubernetes.io/name: {{ include "oceanbase-dashboard.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "oceanbase-dashboard.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "oceanbase-dashboard.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the cleanup cronjob
*/}}
{{- define "oceanbase-dashboard.cleanup.fullname" -}}
{{- printf "%s-cleanup" (include "oceanbase-dashboard.fullname" .) -}}
{{- end -}}

{{/* Keep standalone names stable; embedded releases may scope RBAC by namespace. */}}
{{- define "oceanbase-dashboard.rbacName" -}}
{{- if .Values.rbac.includeNamespace -}}
{{- printf "%s-%s" .Release.Namespace .Release.Name -}}
{{- else -}}
{{- .Release.Name -}}
{{- end -}}
{{- end -}}

{{- define "oceanbase-dashboard.clusterRoleName" -}}
{{- printf "%s-cluster-role" (include "oceanbase-dashboard.rbacName" .) -}}
{{- end -}}

{{- define "oceanbase-dashboard.clusterRoleBindingName" -}}
{{- printf "%s-cluster-role-binding" (include "oceanbase-dashboard.rbacName" .) -}}
{{- end -}}

{{- define "oceanbase-dashboard.prometheusClaim" -}}
{{- default (printf "%s-prometheus-pvc" .Release.Name) .Values.prometheus.persistence.existingClaim -}}
{{- end -}}

{{- define "oceanbase-dashboard.sharedClaim" -}}
{{- default (printf "%s-shared-volume-pvc" .Release.Name) .Values.sharedVolume.persistence.existingClaim -}}
{{- end -}}
