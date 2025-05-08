terraform {
  backend "azurerm" {
    resource_group_name   = "agida-main-uaen-adotf-rg"
    storage_account_name  = "agidamainuaentfsa"
    container_name        = "tfstate"
    key                   = "main/global.tfstate"  
  }
}
