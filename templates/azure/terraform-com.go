package azure

const TerraformAzureCOMMONxMAIN = `# --------------------------------------------------------------------------- #
# --- Input variables
# --------------------------------------------------------------------------- #

variable "config" {
  type = object({
    env                  = string # The current env: dev, qua, or prd, for instance
    location             = string # ex: northeurope
    identity_sub_name    = string # The name of the subscription hosting identity-related resources
    management_sub_name  = string # The name of the subscription hosting resources related to tools & supervision
    environment_sub_name = string # The name of the subscription hosting all the resources needed for this particular environment
    acr_name             = string # The name of the container registry we use
    acr_rg               = string # The resource group the container registry belongs to
    domain_name          = string # The full domain name, something like: companyname.com
    port                 = number # The port the API will listen to
  })
}

# --------------------------------------------------------------------------- #
# --- Computed variables
# --------------------------------------------------------------------------- #

locals {
  subdomain_name = "{{.AppNameLower}}-${var.config.env}.${var.config.domain_name}"
}

# --------------------------------------------------------------------------- #
# --- Providers, associated with subscriptions
# --------------------------------------------------------------------------- #

terraform {
  required_providers {
    azapi = {
      source = "azure/azapi"
    }
  }
}

provider "azurerm" {
  features {}
  alias = "global"
}

data "azurerm_subscriptions" "all" {
  provider = azurerm.global
}

locals {
  identity_subscription = one([
    for s in data.azurerm_subscriptions.all.subscriptions
    : s if s.display_name == var.config.identity_sub_name
  ])
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
  alias           = "identity_sub"
  subscription_id = local.identity_subscription.subscription_id
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
# --- Resources
# --------------------------------------------------------------------------- #

# 1) Resource Group
resource "azurerm_resource_group" "rg" {
  provider = azurerm.environment_sub
  name     = "rg-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  location = var.config.location
  tags     = { environment = var.config.env, application = "{{.AppName}}" }
}

# 2) Log Analytics
resource "azurerm_log_analytics_workspace" "log" {
  provider            = azurerm.environment_sub
  name                = "log-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  location            = var.config.location
  resource_group_name = azurerm_resource_group.rg.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = { environment = var.config.env, application = "{{.AppName}}" }
}

# 3) Container App Environment
resource "azurerm_container_app_environment" "cae" {
  provider                   = azurerm.environment_sub
  name                       = "cae-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  location                   = var.config.location
  resource_group_name        = azurerm_resource_group.rg.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.log.id
  tags                       = { environment = var.config.env, application = "{{.AppName}}" }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.uai.id]
  }
}

# 4) User Assigned Identity (ID)
resource "azurerm_user_assigned_identity" "uai" {
  provider            = azurerm.identity_sub
  name                = "id-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  location            = var.config.location
  resource_group_name = "rg-{resource_ns}-identities"
  tags                = { environment = var.config.env, application = "{{.AppName}}" }
}

# 5) ACR
data "azurerm_container_registry" "acr" {
  provider            = azurerm.management_sub
  name                = var.config.acr_name
  resource_group_name = var.config.acr_rg
}

# 6) AcrPull role for our UAI on ACR
resource "azurerm_role_assignment" "acr_pull" {
  provider             = azurerm.identity_sub
  principal_id         = azurerm_user_assigned_identity.uai.principal_id
  role_definition_name = "AcrPull"
  scope                = data.azurerm_container_registry.acr.id
}

# 7) Key Vault
data "azurerm_client_config" "me" {
  provider = azurerm.global
}

resource "azurerm_key_vault" "kv" {
  provider                   = azurerm.identity_sub
  name                       = "kv-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  location                   = var.config.location
  resource_group_name        = "rg-{resource_ns}-keyvaults"
  tenant_id                  = data.azurerm_client_config.me.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
  purge_protection_enabled   = true
  tags                       = { environment = var.config.env, application = "{{.AppName}}" }
}

# 8.a) Giving ourselves some rights on the KV
resource "azurerm_key_vault_access_policy" "terraform_operator" {
  provider     = azurerm.identity_sub
  key_vault_id = azurerm_key_vault.kv.id
  tenant_id    = data.azurerm_client_config.me.tenant_id
  object_id    = data.azurerm_client_config.me.object_id

  certificate_permissions = ["Get", "List", "Create", "Delete", "Purge", "Recover"]
  secret_permissions      = ["Get", "List", "Set", "Delete"]
}

# 8.b) Giving our UAI the right to use the KV
resource "azurerm_key_vault_access_policy" "uai_kv" {
  provider     = azurerm.identity_sub
  key_vault_id = azurerm_key_vault.kv.id
  tenant_id    = data.azurerm_client_config.me.tenant_id
  object_id    = azurerm_user_assigned_identity.uai.principal_id

  secret_permissions      = ["Get"]
  certificate_permissions = ["Get"]
}

# 9) Self-signed cert in the Key Vault (PKCS#12 / PFX)
resource "azurerm_key_vault_certificate" "self_signed" {
  provider     = azurerm.identity_sub
  name         = "kvcert-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  key_vault_id = azurerm_key_vault.kv.id

  certificate_policy {
    issuer_parameters { name = "Self" }

    key_properties {
      exportable = true
      key_size   = 2048
      key_type   = "RSA"
      reuse_key  = true
    }

    secret_properties {
      content_type = "application/x-pkcs12"
    }

    x509_certificate_properties {
      subject            = "CN=${local.subdomain_name}"
      validity_in_months = 12
      key_usage          = ["digitalSignature", "keyEncipherment"]
      extended_key_usage = ["1.3.6.1.5.5.7.3.1"] # Server Authentication
    }
  }

  depends_on = [
    azurerm_key_vault_access_policy.uai_kv,
    azurerm_key_vault_access_policy.terraform_operator
  ]
}

# 10) Import the KV cert into ACA
resource "azurerm_container_app_environment_certificate" "cert" {
  provider                     = azurerm.environment_sub
  name                         = "caecert-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  container_app_environment_id = azurerm_container_app_environment.cae.id

  certificate_key_vault {
    identity            = azurerm_user_assigned_identity.uai.id
    key_vault_secret_id = azurerm_key_vault_certificate.self_signed.versionless_secret_id
  }

  depends_on = [azurerm_key_vault_access_policy.uai_kv]
}

# 11) Container App (ACA)
resource "azurerm_container_app" "aca" {
  provider                     = azurerm.environment_sub
  name                         = "aca-{resource_ns}-${var.config.env}-{{.AppNameLower}}"
  resource_group_name          = azurerm_resource_group.rg.name
  container_app_environment_id = azurerm_container_app_environment.cae.id
  revision_mode                = "Single"

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.uai.id]
  }

  registry {
    server   = data.azurerm_container_registry.acr.login_server
    identity = azurerm_user_assigned_identity.uai.id
  }

  template {
    container {
      name   = "{{.AppNameKebab}}-api"
      image  = "${var.config.acr_name}.azurecr.io/{{.AppNameKebab}}-api:latest"
      cpu    = 0.5
      memory = "1Gi"

      env {
        name  = "ENV"
        value = var.config.env
      }
      env {
        name  = "OIDC_TENANT_ID"
        value = var.config.domain_name
      }
      env {
        name  = "OIDC_AUDIENCE"
        value = "api://${var.config.domain_name}/app-{resource_ns}-{{.AppNameLower}}"
      }
    }
  }

  ingress {
    external_enabled = true
    target_port      = var.config.port

    traffic_weight {
      label           = "stable"
      percentage      = 100
      latest_revision = true
    }
  }

  lifecycle {
    # CI/CD updates the running image; Terraform should not revert it on next apply.
    ignore_changes = [
      template[0].container[0].image
    ]
  }

  tags = { environment = var.config.env, application = "{{.AppName}}" }
}

# 12) Binding of the subdomain to the ACA app (via azapi PATCH), using the CAE cert defined here
# resource "azapi_update_resource" "aca_subdomain" {
#   type        = "Microsoft.App/containerApps@2023-05-01"
#   resource_id = azurerm_container_app.aca.id

#   body = {
#     properties = {
#       configuration = {
#         ingress = {
#           customDomains = [
#             {
#               name          = local.subdomain_name
#               bindingType   = "SniEnabled"
#               certificateId = azurerm_container_app_environment_certificate.cert.id
#             }
#           ]
#         }
#       }
#     }
#   }

#   depends_on = [
#     azurerm_container_app.aca,
#     azurerm_container_app_environment_certificate.cert
#   ]
# }

# ----------------------- Outputs utiles -----------------------

output "aca_api_base_url" {
  description = "Base HTTPS URL of the ACA API"
  value       = "https://${azurerm_container_app.aca.ingress[0].fqdn}"
}

# output "custom_domain" {
#   description = "ACA's full custom domain name"
#   value       = local.subdomain_name
# }
`
