package templates

const AppLocal = `# each backend pod
kind: Deployment
apiVersion: apps/v1
metadata:
  name: {{.AppName}}-api-depl
  namespace: {{.AppName}}-local

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

---

# the load balancer routing to the backend pods
kind: Service
apiVersion: v1
metadata:
  name: {{.AppName}}-api-lb
  namespace: {{.AppName}}-local

spec:
  selector:
    app: {{.AppName}}-api
  ports:
    - name: http
      port: {{.API.Port}} # load balancer port (external) - set by Aldev
      targetPort: 55555 # API port (internal) - should not be changed
  type: LoadBalancer
`

const AppLocalFrontContainer=`---

# the frontend pod
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.AppName}}-web-depl
  namespace: {{.AppName}}-local
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
