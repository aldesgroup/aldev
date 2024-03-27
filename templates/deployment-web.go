package templates

// namespace: {{.AppName}}-local
const Web = `# the frontend pod
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.AppName}}-web-depl
  labels:
    app: {{.AppName}}-web

spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.AppName}}-web
  template:
    metadata:
      labels:
        app: {{.AppName}}-web
    spec:
      containers:
        - name: {{.AppName}}-web
          image: {{.AppName}}-web-image
          env:
            - name: VITE_CLIENT_PORT
              value: '{{.Web.Port}}'
          ports:
            - containerPort: 3000
`
