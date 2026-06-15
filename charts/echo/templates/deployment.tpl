apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "echo.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "echo.labels" . | nindent 4 }}
  {{- with .Values.deploymentAnnotations }}
  # Workload-level annotations — e.g. a Stakater Reloader annotation, which must
  # sit on the Deployment (not the pod) to roll it when a referenced object changes.
  annotations:
    {{- tpl (toYaml .) $ | nindent 4 }}
  {{- end }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "echo.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "echo.labels" . | nindent 8 }}
        {{- with .Values.podLabels }}
        {{- tpl (toYaml .) $ | nindent 8 }}
        {{- end }}
      {{- with .Values.podAnnotations }}
      annotations:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "echo.serviceAccountName" . }}
      automountServiceAccountToken: {{ .Values.serviceAccount.automount }}
      {{- with .Values.priorityClassName }}
      priorityClassName: {{ tpl . $ | quote }}
      {{- end }}
      terminationGracePeriodSeconds: {{ .Values.terminationGracePeriodSeconds }}
      securityContext:
        {{- tpl (toYaml .Values.podSecurityContext) $ | nindent 8 }}
      containers:
        - name: echo
          image: {{ include "echo.image" . | quote }}
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          securityContext:
            {{- tpl (toYaml .Values.securityContext) $ | nindent 12 }}
          env:
            - name: ECHO_HTTP_PORT
              value: {{ .Values.config.httpPort | quote }}
            - name: ECHO_METRICS_ENABLED
              value: {{ .Values.config.metricsEnabled | quote }}
            - name: ECHO_METRICS_ADDR
              value: {{ printf ":%v" .Values.config.metricsPort | quote }}
            - name: ECHO_LOG_LEVEL
              value: {{ .Values.config.logLevel | quote }}
            - name: ECHO_LOG_FORMAT
              value: {{ .Values.config.logFormat | quote }}
            - name: ECHO_DISABLE_REQUEST_LOGS
              value: {{ .Values.config.disableRequestLogs | quote }}
            - name: ECHO_BACK_TO_CLIENT
              value: {{ .Values.config.echoBackToClient | quote }}
            - name: ECHO_MAX_BODY_BYTES
              value: {{ .Values.config.maxBodyBytes | int64 | quote }}
            - name: ECHO_WS_ENABLED
              value: {{ .Values.config.wsEnabled | quote }}
            {{- with .Values.config.wsAllowedOrigins }}
            - name: ECHO_WS_ALLOWED_ORIGINS
              value: {{ join "," . | quote }}
            {{- end }}
            {{- with .Values.config.trustedProxies }}
            - name: ECHO_TRUSTED_PROXIES
              value: {{ join "," . | quote }}
            {{- end }}
            {{- if .Values.config.kubernetes }}
            # Pod/node identity via the Downward API; surfaced in responses as a kubernetes object.
            - name: ECHO_KUBERNETES
              value: "true"
            - name: ECHO_POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: ECHO_POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: ECHO_POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: ECHO_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            {{- end }}
            {{- range $k, $v := .Values.env }}
            - name: {{ $k }}
              value: {{ tpl (toString $v) $ | quote }}
            {{- end }}
            {{- with .Values.extraEnv }}
            {{- tpl (toYaml .) $ | nindent 12 }}
            {{- end }}
          {{- with .Values.envFrom }}
          envFrom:
            {{- tpl (toYaml .) $ | nindent 12 }}
          {{- end }}
          ports:
            - name: http
              containerPort: {{ .Values.config.httpPort }}
              protocol: TCP
            {{- if .Values.config.metricsEnabled }}
            - name: metrics
              containerPort: {{ .Values.config.metricsPort }}
              protocol: TCP
            {{- end }}
          {{- with .Values.startupProbe }}
          startupProbe:
            {{- tpl (toYaml .) $ | nindent 12 }}
          {{- end }}
          {{- with .Values.livenessProbe }}
          livenessProbe:
            {{- tpl (toYaml .) $ | nindent 12 }}
          {{- end }}
          {{- with .Values.readinessProbe }}
          readinessProbe:
            {{- tpl (toYaml .) $ | nindent 12 }}
          {{- end }}
          {{- with .Values.resources }}
          resources:
            {{- tpl (toYaml .) $ | nindent 12 }}
          {{- end }}
          {{- with .Values.volumeMounts }}
          volumeMounts:
            {{- tpl (toYaml .) $ | nindent 12 }}
          {{- end }}
      {{- with .Values.volumes }}
      volumes:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      {{- with .Values.topologySpreadConstraints }}
      topologySpreadConstraints:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
