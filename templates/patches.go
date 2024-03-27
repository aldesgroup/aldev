package templates

const NoWebContainerPatch = `apiVersion: apps/v1
$patch: delete
kind: Deployment
metadata:
  name: {{.AppName}}-web-depl`
