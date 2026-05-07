package azure

const TerraformAzureAPIMxENV = `# --------------------------------------------------------------------------- #
# --- Main Terraform config to update the APIM configuration for the %[1]s env
# --------------------------------------------------------------------------- #
locals {
  config = {
    # The current env: dev, qua, or prd, for instance
    env = "%[1]s"
    # The name of the subscription hosting resources related to tools & supervision
    management_sub_name = "{management_sub_name}"
    # The name of the subscription hosting all the resources needed for this particular environment
    environment_sub_name = "{environment_sub_name}"
    # The name of the API manager we use
    apim_name = "{apim_name}"
    # The resource group the API manager belongs to
    apim_rg = "{apim_rg}"
  }
}

module "apim_%[1]s" {
  source = "../"
  config = local.config
}

`
