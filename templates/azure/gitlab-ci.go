package azure

const GitlabAzureCIxCDxCONF = `stages: [ build, deploy_test, deploy_prod ]

# --------------------------------------------------------------------------- #
# --- This is the default behaviour of each step
# --------------------------------------------------------------------------- #

default:
  id_tokens:
    AZURE_ID_TOKEN:
      aud: api://AzureADTokenExchange
  before_script:
  - |
    az login                                    \
      --service-principal                       \
      --tenant                 ${ARM_TENANT_ID} \
      --username               ${ARM_CLIENT_ID} \
      --federated-token        $AZURE_ID_TOKEN  \
      --output                 none             \
      --allow-no-subscriptions

# --------------------------------------------------------------------------- #
# --- Defining reusable instructions
# --------------------------------------------------------------------------- #

.azure_set_mgmt_sub: &azure_set_mgmt_sub
  - az account set --subscription {management_sub_name}

.acr_podman_login: &acr_podman_login
  - ACR_TOKEN=$(az acr login --name {acr_name} --expose-token --query accessToken -o tsv)
  - echo $ACR_TOKEN | podman login {acr_name}.azurecr.io -u 00000000-0000-0000-0000-000000000000 --password-stdin

# --------------------------------------------------------------------------- #
# --- Building & deploying on {env-SANDBOX}, with each new commit on main
# --------------------------------------------------------------------------- #

build_{env-SANDBOX}:
  stage: build
  tags: [ {resource_ns}, {env-SANDBOX} ]
  script:
  - *azure_set_mgmt_sub
  - *acr_podman_login
  - export IMAGE={acr_name}.azurecr.io/{{.AppNameKebab}}-api:${CI_COMMIT_SHORT_SHA}
  - podman build -t $IMAGE -f deploy/Containerfile --build-arg ENV={env-SANDBOX} . 
  - podman push $IMAGE
  rules:
  - if: '$CI_COMMIT_BRANCH == "main"'
    when: on_success

deploy_{env-SANDBOX}:
  stage: deploy_test
  tags: [ {resource_ns}, {env-SANDBOX} ]
  needs: [ build_{env-SANDBOX} ]
  script:
  - |
    az containerapp update                                                    \
      --subscription   {sub-SANDBOX}                                           \
      --name           aca-{resource_ns}-{env-SANDBOX}-{{.AppNameLower}}                                 \
      --resource-group rg-{resource_ns}-{env-SANDBOX}-{{.AppNameLower}}                                  \
      --container-name {{.AppNameKebab}}-api                                         \
      --image          {acr_name}.azurecr.io/{{.AppNameKebab}}-api:${CI_COMMIT_SHORT_SHA} \
      --output         none
  environment:
    name: {env-SANDBOX}
  rules:
  - if: '$CI_COMMIT_BRANCH == "main"'
    when: on_success

# --------------------------------------------------------------------------- #
# --- Building & deploying on {env-STAGING} & {env-PRODUCTION}, with each new tag
# --------------------------------------------------------------------------- #

build_{env-STAGING}_n_{env-PRODUCTION}:
  stage: build
  tags: [ {resource_ns}, {env-STAGING} ]
  script:
  - *azure_set_mgmt_sub
  - *acr_podman_login
  - export IMAGE_STAGING={acr_name}.azurecr.io/{{.AppNameKebab}}-api:${CI_COMMIT_TAG}-{env-STAGING}
  - podman build -t $IMAGE_STAGING -f deploy/Containerfile --build-arg ENV={env-STAGING} .
  - podman push $IMAGE_STAGING
  - export IMAGE_PRODUCTION={acr_name}.azurecr.io/{{.AppNameKebab}}-api:${CI_COMMIT_TAG}-{env-PRODUCTION}
  - podman build -t $IMAGE_PRODUCTION -f deploy/Containerfile --build-arg ENV={env-PRODUCTION} .
  - podman push $IMAGE_PRODUCTION
  rules:
  - if: '$CI_COMMIT_TAG'
    when: on_success

deploy_{env-STAGING}:
  stage: deploy_test
  tags: [ {resource_ns}, {env-STAGING} ]
  needs: [ build_{env-STAGING}_n_{env-PRODUCTION} ]
  script:
  - |
    az containerapp update                                                  \
      --subscription   {sub-STAGING}                                         \
      --name           aca-{resource_ns}-{env-STAGING}-{{.AppNameLower}}                               \
      --resource-group rg-{resource_ns}-{env-STAGING}-{{.AppNameLower}}                                \
      --container-name {{.AppNameKebab}}-api                                       \
      --image          {acr_name}.azurecr.io/{{.AppNameKebab}}-api:${CI_COMMIT_TAG}-{env-STAGING} \
      --output         none
  environment:
    name: {env-STAGING}
  rules:
  - if: '$CI_COMMIT_TAG'
    when: on_success

deploy_{env-PRODUCTION}:
  stage: deploy_prod
  tags: [ {resource_ns}, {env-STAGING} ]
  needs: [ deploy_{env-STAGING} ]
  script:
  - |
    az containerapp update                                                  \
      --subscription   {sub-PRODUCTION}                                         \
      --name           aca-{resource_ns}-{env-PRODUCTION}-{{.AppNameLower}}                               \
      --resource-group rg-{resource_ns}-{env-PRODUCTION}-{{.AppNameLower}}                                \
      --container-name {{.AppNameKebab}}-api                                       \
      --image          {acr_name}.azurecr.io/{{.AppNameKebab}}-api:${CI_COMMIT_TAG}-{env-PRODUCTION} \
      --output         none
  environment:
    name: {env-PRODUCTION}
  rules:
  - if: '$CI_COMMIT_TAG'
    when: manual

# # ----------------------------
# # QUA: tag last-known-good after successful QUA deploy
# # ----------------------------
# mark_qua_lkg:
#   stage: deploy
#   needs: [ deploy_qua ]
#   script:
#   - *azure_set_mgmt_sub
#   - az acr import \ --name "{acr_name}" \ --source "{acr_name}.azurecr.io/${IMAGE_REPO}:${CI_COMMIT_TAG}" \ --image "${IMAGE_REPO}:qua-lkg" \ --force
#   rules:
#   - if: '$CI_COMMIT_TAG'
#     when: on_success

# # ----------------------------
# # PRD: manual deploy, reusing the exact tag built/deployed in QUA
# # ----------------------------
# deploy_prd:
#   stage: deploy
#   needs: [ mark_qua_lkg ]
#   script:
#   - az account set --subscription "subs-{resource_ns}-prd"
#   - export ENV="prd"
#   - export ACA_RG="rg-{resource_ns}-${ENV}-${APP}"
#   - export ACA_NAME="aca-{resource_ns}-${ENV}-${APP}"
#   - export IMAGE_TAG="${CI_COMMIT_TAG}"
#   - export IMAGE="{acr_name}.azurecr.io/${IMAGE_REPO}:${IMAGE_TAG}"
#   - *deploy_aca_image
#   environment:
#     name: prd
#   when: manual
#   allow_failure: false
#   rules:
#   - if: '$CI_COMMIT_TAG'
#     when: manual

# # ----------------------------
# # PRD: tag last-known-good after successful manual PRD deploy
# # ----------------------------
# mark_prd_lkg:
#   stage: deploy
#   needs: [ deploy_prd ]
#   script:
#   - *azure_set_mgmt_sub
#   - az acr import \ --name "{acr_name}" \ --source "{acr_name}.azurecr.io/${IMAGE_REPO}:${CI_COMMIT_TAG}" \ --image "${IMAGE_REPO}:prd-lkg" \ --force
#   rules:
#   - if: '$CI_COMMIT_TAG'
#     when: on_success

# # ----------------------------
# # Manual rollback (any env, any tag, including qua-lkg / prd-lkg)
# # Trigger with variables:
# #   ROLLBACK_ENV={env-SANDBOX}|qua|prd
# #   ROLLBACK_TAG=<sha>|vX.Y.Z|qua-lkg|prd-lkg
# # ----------------------------
# rollback:
#   stage: deploy
#   variables:
#     ROLLBACK_ENV:
#       value: "qua"
#       options: [ {env-SANDBOX}", "qua", "prd ]
#       description: "Target environment to roll back."
#     ROLLBACK_TAG:
#       value: "qua-lkg"
#       options: [ qua-lkg", "prd-lkg ]
#       description: "Tag to deploy. Use qua-lkg/prd-lkg for last known good, or type a specific tag (vX.Y.Z / sha) if allowed by your GitLab version."
#   script:
#   - |
#     if [ -z "$ROLLBACK_ENV ] || [ -z "$ROLLBACK_TAG ]; then
#       echo "ERROR: set ROLLBACK_ENV ({env-SANDBOX}|qua|prd) and ROLLBACK_TAG (e.g. qua-lkg, prd-lkg, v1.2.3)"
#       exit 1
#     fi

#   - az account set --subscription "subs-{resource_ns}-${ROLLBACK_ENV}"
#   - export ENV="$ROLLBACK_ENV"
#   - export ACA_RG="rg-{resource_ns}-${ENV}-${APP}"
#   - export ACA_NAME="aca-{resource_ns}-${ENV}-${APP}"
#   - export IMAGE="{acr_name}.azurecr.io/${IMAGE_REPO}:${ROLLBACK_TAG}"
#   - *deploy_aca_image
#   environment:
#     name: rollback
#   when: manual
#   allow_failure: false
#   rules:
#   - when: manual
`
