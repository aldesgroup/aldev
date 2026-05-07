package azure

const TerraformAzureAPIMxCOMMON = `# --------------------------------------------------------------------------- #
# --- APIM integration for the {{.AppName}} API
# ---
# --- This file is meant to be re-applied by CI/CD whenever the API doc
# --- (eg data/api-doc.yaml) changes, 
# --- so that APIM always reflects the current API contract.
# --------------------------------------------------------------------------- #

# --------------------------------------------------------------------------- #
# --- Input variables
# --------------------------------------------------------------------------- #

variable "config" {
  type = object({
    env                  = string # The current env: dev, qua, or prd, for instance
    management_sub_name  = string # The name of the subscription hosting resources related to tools & supervision
    environment_sub_name = string # The name of the subscription hosting all the resources needed for this particular environment
    apim_name            = string # The name of the API manager we use
    apim_rg              = string # The resource group the API manager belongs to
  })
}

# --------------------------------------------------------------------------- #
# --- Subscriptions
# --------------------------------------------------------------------------- #

provider "azurerm" {
  features {}
  alias = "global"
}

data "azurerm_subscriptions" "all" {
  provider = azurerm.global
}

locals {
  management_subscription = one([
    for s in data.azurerm_subscriptions.all.subscriptions
    : s if s.display_name == var.config.management_sub_name
  ])
  environment_subscription = one([
    for s in data.azurerm_subscriptions.all.subscriptions
    : s if s.display_name == var.config.environment_sub_name
  ])
}

provider "azurerm" {
  features {}
  alias           = "management_sub"
  subscription_id = local.management_subscription.subscription_id
}

provider "azurerm" {
  features {}
  alias           = "environment_sub"
  subscription_id = local.environment_subscription.subscription_id
}

# --------------------------------------------------------------------------- #
# --- Data sources
# --------------------------------------------------------------------------- #

# The APIM instance (lives in the management subscription)
data "azurerm_api_management" "apim" {
  provider            = azurerm.management_sub
  name                = var.config.apim_name
  resource_group_name = var.config.apim_rg
}

# The ACA container app for current environment (lives in the current env sub)
data "azurerm_container_app" "aca" {
  provider            = azurerm.environment_sub
  name                = "aca-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  resource_group_name = "rg-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
}

# --------------------------------------------------------------------------- #
# --- APIM Backend → points to the ACA app ingress
# --------------------------------------------------------------------------- #

resource "azurerm_api_management_backend" "backend" {
  provider            = azurerm.management_sub
  name                = "backend-{{.AppNameLower}}-${var.config.env}"
  resource_group_name = var.config.apim_rg
  api_management_name = data.azurerm_api_management.apim.name
  protocol            = "http"
  url                 = "https://${data.azurerm_container_app.aca.ingress[0].fqdn}"
}

# --------------------------------------------------------------------------- #
# --- APIM API → imported from the OpenAPI spec
# ---
# --- Re-running 'terraform apply' after a spec change will push the updated
# --- contract to APIM automatically (CI/CD trigger point).
# --------------------------------------------------------------------------- #

resource "azurerm_api_management_api" "api" {
  provider              = azurerm.management_sub
  name                  = "{{.AppNameLower}}-${var.config.env}"
  resource_group_name   = var.config.apim_rg
  api_management_name   = data.azurerm_api_management.apim.name
  revision              = "1"
  display_name          = "{{.AppName}} (${var.config.env})"
  path                  = "{{.AppNameLower}}-${var.config.env}"
  protocols             = ["https"]
  subscription_required = false
  description           = "{{.AppDesc}}"

  # Loaded from the project's API doc: any change to that file will be
  # detected by Terraform and pushed to APIM on the next apply.
  import {
    content_format = "openapi"
    content_value  = file("../../../../{{.API.DocPath}}")
  }
}

# --------------------------------------------------------------------------- #
# --- APIM API Policy → routes every call to the ACA backend
# --------------------------------------------------------------------------- #

resource "azurerm_api_management_api_policy" "api_policy" {
  provider            = azurerm.management_sub
  api_name            = azurerm_api_management_api.api.name
  api_management_name = data.azurerm_api_management.apim.name
  resource_group_name = var.config.apim_rg

  xml_content = <<-XML
    <policies>
      <inbound>
        <base />
        <set-backend-service backend-id="${azurerm_api_management_backend.backend.name}" />
      </inbound>
      <backend><base /></backend>
      <outbound><base /></outbound>
      <on-error><base /></on-error>
    </policies>
  XML
}
`
