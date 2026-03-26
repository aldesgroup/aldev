package templates

const TerraformAzureBACKEND = `# --------------------------------------------------------------------------- #
# -- Terraform state storage fdor the resources described in this folder
# --------------------------------------------------------------------------- #

terraform {
  backend "azurerm" {
    resource_group_name  = "rg-{resource_ns}-tfstates"
    storage_account_name = "st{resource_ns}tfstates"
    container_name       = "sc-{resource_ns}-tfstates"
    key                  = "tfstates.{{.AppNameLower}}.%s"
  }
}
`

const TerraformAzureENVxMAIN = `# --------------------------------------------------------------------------- #
# --- Main Terraform config to create the whole infra for the {env} environment
# --------------------------------------------------------------------------- #
locals {
  config = {
    # The current env: dev, qua, or prd, for instance
    env = "%[1]s"
    # ex: northeurope
    location = "{location}"
    # The name of the subscription hosting identity-related resources
    identity_sub_name = "{identity_sub_name}"
    # The name of the subscription hosting resources related to tools & supervision
    management_sub_name = "{management_sub_name}"
    # The name of the subscription hosting all the resources needed for this particular environment
    environment_sub_name = "{environment_sub_name}"
    # The name of the container registry we use
    acr_name = "{acr_name}"
    # The resource group the container registry belongs to
    acr_rg = "{acr_rg}"
    # The domain name, something like: companyname.com
    domain_name = "{domain_name}"
  }
}

module "infra_%[1]s" {
  source = "../"
  config = local.config
}
`

const TerraformAzureGLOBALxMAIN = `# --------------------------------------------------------------------------- #
# --- Main Terraform config to the global resources, like the app reg
# --------------------------------------------------------------------------- #

# --------------------------------------------------------------------------- #
# --- App Registration = definition of the app
# - Defines a name
# - Scopes (access_as_user)
# - The audience (api://...)
# - Expose permissions
# --------------------------------------------------------------------------- #

resource "azuread_application" "app_{resource_ns}_{{.AppNameLower}}" {
  display_name = "app-{resource_ns}-{{.AppNameLower}}"

  # IMPORTANT : URI lisible mais conforme à la policy du tenant
  identifier_uris = ["api://{domain_name}/app-{resource_ns}-{{.AppNameLower}}"]

  api {
    requested_access_token_version = 2

    oauth2_permission_scope {
      admin_consent_description  = "Allow access to {{.AppName}} API"
      admin_consent_display_name = "Access {{.AppName}} API"
      enabled                    = true
      id                         = "{oauth2_scope_guid}"
      type                       = "User"
      value                      = "access_as_user"
    }
  }
}

# --------------------------------------------------------------------------- #
# --- Service principal = operationnal instance of the app
# - Usable identity
# - Visible in Enterprise Apps
# - Can receive RBAC roles
# --------------------------------------------------------------------------- #

resource "azuread_service_principal" "sp_{resource_ns}_{{.AppNameLower}}" {
  client_id = azuread_application.app_{resource_ns}_{{.AppNameLower}}.client_id
}
`
