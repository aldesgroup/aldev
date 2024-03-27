package templates

// namespace: {{.AppName}}-local
const LB = `# the load balancer routing to the backend pods
apiVersion: v1
kind: Service
metadata:
  name: {{.AppName}}-api-lb

spec:
  selector:
    app: {{.AppName}}-api
  ports:
    - name: http
      port: {{.API.Port}} # load balancer port (external) - set by Aldev
      targetPort: 55555 # API port (internal) - should not be changed
  type: LoadBalancer
`
