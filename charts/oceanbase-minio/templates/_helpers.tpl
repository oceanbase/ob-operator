{{- define "obminio.labels" -}}
app.kubernetes.io/name: oceanbase-minio
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
{{- define "obminio.secret" -}}
{{- default (printf "%s-minio-root" .Release.Name) .Values.rootExistingSecret -}}
{{- end -}}
{{- define "obminio.scheduling" -}}
{{- with .Values.nodeSelector }}nodeSelector: {{ toYaml . | nindent 2 }}{{ end }}
{{- with .Values.tolerations }}
tolerations: {{ toYaml . | nindent 2 }}{{ end }}
{{- end -}}
