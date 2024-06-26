package templates

const GitlabCI = `# Generated by aldev, do not edit!
stages:
- build-backend
- build-frontend
- deploy-local
- deploy-sandbox
- deploy-staging

build-backend:
  stage: build-backend
  script:
    - docker build -t backend-image /path/to/backend
    - docker push backend-image
  only:
    - branches
  # other configuration for building the backend Docker image

build-frontend:
  stage: build-frontend
  script:
    - docker build -t frontend-image /path/to/frontend
    - docker push frontend-image
  only:
    - branches
  # other configuration for building the frontend Docker image


deploy-sandbox:
  stage: deploy-sandbox
  script:
    - kubectl apply -k deploy/overlays/sandbox
  only:
    - master`
