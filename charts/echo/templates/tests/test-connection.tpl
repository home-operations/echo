apiVersion: v1
kind: Pod
metadata:
  name: {{ include "echo.fullname" . }}-test-connection
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "echo.labels" . | nindent 4 }}
  annotations:
    helm.sh/hook: test
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
spec:
  restartPolicy: Never
  securityContext:
    {{- toYaml .Values.podSecurityContext | nindent 4 }}
  containers:
    - name: curl
      image: {{ include "echo.testImage" . | quote }}
      imagePullPolicy: {{ .Values.tests.image.pullPolicy }}
      securityContext:
        {{- toYaml .Values.securityContext | nindent 8 }}
      command:
        - /bin/sh
        - -c
        - |
          set -eu
          url="http://{{ include "echo.fullname" . }}:{{ .Values.service.port }}/helm-test?ok=1"
          echo "GET ${url}"
          body="$(curl -fsS "${url}")"
          echo "${body}"
          echo "${body}" | grep -q '"path":"/helm-test"'
          echo "${body}" | grep -q '"method":"GET"'
