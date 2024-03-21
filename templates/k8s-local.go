package templates

const K8sLocal = `# each backend pod
kind: Deployment
apiVersion: apps/v1
metadata:
  name: {{.AppName}}-back
  namespace: {{.AppName}}-local

spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.AppName}}-local
  template:
    metadata:
      labels:
        app: {{.AppName}}-local
    spec:
      containers:
        - name: {{.AppName}}-local
          image: {{.AppName}}-local-image
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
  name: {{.AppName}}-local-service
  namespace: {{.AppName}}-local

spec:
  selector:
    app: {{.AppName}}-local
  ports:
    - name: http
      port: 24243 # load balancer port (external)
      targetPort: 55555 # API port (internal)
  type: LoadBalancer
`
