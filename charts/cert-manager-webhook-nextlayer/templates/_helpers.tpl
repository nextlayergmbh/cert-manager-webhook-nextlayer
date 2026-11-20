{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cert-manager-webhook-nextlayer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Namespace resolution order:
  1. CLI (-n / --namespace), if not "default"
  2. .Values.namespace
  3. "default"
*/}}
{{- define "cert-manager-webhook-nextlayer.namespace" -}}
{{- if ne .Release.Namespace "default" -}}
{{- .Release.Namespace -}}
{{- else -}}
{{- .Values.namespace | default "default" -}}
{{- end -}}
{{- end -}}

{{/*
Port the Service (and APIService) exposes.
*/}}
{{- define "cert-manager-webhook-nextlayer.servicePort" -}}
{{- .Values.service.port | default 443 | int -}}
{{- end -}}

{{/*
Port the webhook container listens on (non-privileged by default).
*/}}
{{- define "cert-manager-webhook-nextlayer.securePort" -}}
{{- .Values.webhook.securePort | default 8443 | int -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cert-manager-webhook-nextlayer.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cert-manager-webhook-nextlayer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "cert-manager-webhook-nextlayer.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "cert-manager-webhook-nextlayer.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-nextlayer.rootCAIssuer" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-nextlayer.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-nextlayer.rootCACertificate" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-nextlayer.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-nextlayer.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "cert-manager-webhook-nextlayer.fullname" .) }}
{{- end -}}

{{/*
Selector labels.
Kept identical to the existing selector (app + release), because a
Deployment's spec.selector is immutable. Changing it would break upgrades.
*/}}
{{- define "cert-manager-webhook-nextlayer.selectorLabels" -}}
app: {{ include "cert-manager-webhook-nextlayer.name" . }}
release: {{ .Release.Name }}
{{- end -}}

{{/*
Common labels: legacy labels plus the recommended app.kubernetes.io/* set.
*/}}
{{- define "cert-manager-webhook-nextlayer.labels" -}}
{{ include "cert-manager-webhook-nextlayer.selectorLabels" . }}
chart: {{ include "cert-manager-webhook-nextlayer.chart" . }}
heritage: {{ .Release.Service }}
app.kubernetes.io/name: {{ include "cert-manager-webhook-nextlayer.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ include "cert-manager-webhook-nextlayer.chart" . }}
{{- end -}}
