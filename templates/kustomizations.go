package templates

const NewNamespace = `apiVersion: v1
kind: Namespace
metadata:
  name: {{.AppName}}-%s`

const KustomizationBase = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

# Specify the resources in the base
resources:
  - {{.AppName}}-api-.yaml
  - {{.AppName}}-api-lb.yaml
  - {{.AppName}}-cm.yaml
  - {{.AppName}}-web.yaml`

const KustomizationOverlay = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.AppName}}-%s
resources:
  - ../../base
  - namespace-%s.yaml`

// const KustomizationLocal = kustomizationOverlay + "\n" +
// 	"patches:" + "\n" +
// 	"  - path: patch-no-web-container.yaml"

// const KustomizationDev = kustomizationOverlay
// const KustomizationSandbox = kustomizationOverlay