{{- if .Values.httpRoute.enabled }}
apiVersion: {{ .Values.httpRoute.apiVersion | default "gateway.networking.k8s.io/v1" }}
kind: {{ .Values.httpRoute.kind | default "HTTPRoute" }}
metadata:
  name: {{ include "echo.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "echo.labels" . | nindent 4 }}
    {{- with .Values.httpRoute.labels }}
    {{- tpl (toYaml .) $ | nindent 4 }}
    {{- end }}
  {{- with .Values.httpRoute.annotations }}
  annotations:
    {{- tpl (toYaml .) $ | nindent 4 }}
  {{- end }}
spec:
  {{- with .Values.httpRoute.parentRefs }}
  parentRefs:
    {{- tpl (toYaml .) $ | nindent 4 }}
  {{- end }}
  {{- with .Values.httpRoute.hostnames }}
  hostnames:
    {{- tpl (toYaml .) $ | nindent 4 }}
  {{- end }}
  rules:
    - matches:
        {{- tpl (toYaml .Values.httpRoute.matches) $ | nindent 8 }}
      backendRefs:
        - name: {{ include "echo.fullname" . }}
          port: {{ .Values.service.port }}
{{- end }}
