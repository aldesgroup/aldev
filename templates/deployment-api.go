package templates

// namespace: {{.AppName}}-local
const API = `# each backend pod
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.AppName}}-api-depl

spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.AppName}}-api
  template:
    metadata:
      labels:
        app: {{.AppName}}-api
    spec:
      containers:
        - name: {{.AppName}}-api
          image: {{.AppName}}-api-image
          args:
            - "-config=config/config.yaml"
          volumeMounts:
          - name: config-volume
            mountPath: /api/config
      volumes:
      - name: config-volume
        configMap:
          name: {{.AppName}}-configmap
`
