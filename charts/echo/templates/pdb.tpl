{{- if .Values.podDisruptionBudget.enabled }}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "echo.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "echo.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "echo.selectorLabels" . | nindent 6 }}
  {{- if .Values.podDisruptionBudget.maxUnavailable }}
  maxUnavailable: {{ .Values.podDisruptionBudget.maxUnavailable }}
  {{- else }}
  minAvailable: {{ .Values.podDisruptionBudget.minAvailable }}
  {{- end }}
{{- end }}
