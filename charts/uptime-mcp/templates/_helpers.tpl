{{/*
Expand the name of the chart.
*/}}
{{- define "uptime-mcp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "uptime-mcp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "uptime-mcp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "uptime-mcp.labels" -}}
helm.sh/chart: {{ include "uptime-mcp.chart" . }}
{{ include "uptime-mcp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "uptime-mcp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "uptime-mcp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "uptime-mcp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "uptime-mcp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Container args assembled from config values.
*/}}
{{- define "uptime-mcp.args" -}}
- -transport={{ .Values.config.transport }}
- -listen={{ .Values.config.listen }}
{{- with .Values.config.uptimeUrl }}
- -uptime-url={{ . }}
{{- end }}
{{- with .Values.config.resourceUrl }}
- -resource-url={{ . }}
{{- end }}
{{- with .Values.config.clientId }}
- -client-id={{ . }}
{{- end }}
{{- with .Values.config.logLevel }}
- -log-level={{ . }}
{{- end }}
{{- range .Values.config.extraArgs }}
- {{ . }}
{{- end }}
{{- end }}
