module "sys_resource_group" {
    source                          = "../../Resources/resource-group"
    name                            = "${var.project_name}-${var.environment}-${var.region_short}-sys-rg" 
    location                        = var.region
}

/*
*/

module "ready_azure_function_app" {
    source                          = "../../Resources/function-app"
    function_app_name               = "${var.project_name}-${var.environment}-${var.region_short}-sysReady-fapp" 
    function_resource_group_name    = module.sys_resource_group.name
    storage_account_name            = "${var.project_name}${var.environment}${var.region_short}sysfardystg"
    location                        = var.region
    app_service_plan_name           = "${var.project_name}-${var.environment}-${var.region_short}-sysReady-fasp" 
    app_service_plan_sku_name       = var.const_app_service_plan_sku_name_b1
    app_service_plan_os_type        = var.const_app_service_plan_os_type_windows
    storage_sku                     = var.const_storage_sku_Standard
    storage_resource_group_name     = module.sys_resource_group.name
    storage_replication_type        = var.const_storage_replication_type_LRS
    function_runtime_version        = var.const_function_runtime_version_4
    function_dotnet_version         = var.const_function_dotnet_version_v8
    function_subnet_id              = var.sysfapp_subnet_cidr

    depends_on = [ module.sys_resource_group ]
}

/*
*/

module "ready_azure_function_app_pep" {
    source                          = "../../Resources/private-endpoint"
    name                            = "${var.project_name}-${var.environment}-${var.region_short}-sysfappReady-pep"
    location                        = var.region
    resource_group_name             = module.sys_resource_group.name
    subnet_id                       = var.sysfapp_pep_subnet_cidr
    target_resource_id              = module.ready_azure_function_app.function_resource_id
    subresource_names               = [var.const_resource_name_sites]
    private_dns_zones               = var.private_webdns_zones
    depends_on = [ module.ready_azure_function_app ]
}


/* 
module "ready_azure_functioniso_app" {
    source                          = "../../Resources/function-app-netcore8iso"
    function_app_name               = "${var.project_name}-${var.environment}-${var.region_short}-sysReadyiso-fapp" 
    function_resource_group_name    = module.sys_resource_group.name
    storage_account_name            = "${var.project_name}${var.environment}${var.region_short}sysfaisostg"
    location                        = var.region
    app_service_plan_name           = "${var.project_name}-${var.environment}-${var.region_short}-sysReadyiso-fasp" 
    app_service_plan_sku_name       = var.const_app_service_plan_sku_name_b1
    app_service_plan_os_type        = var.const_app_service_plan_os_type_windows
    storage_sku                     = var.const_storage_sku_Standard
    storage_resource_group_name     = module.sys_resource_group.name
    storage_replication_type        = var.const_storage_replication_type_LRS
    function_runtime_version        = var.const_function_runtime_version_4
    function_dotnet_version         = var.const_function_dotnet_version_v8
    function_subnet_id              = var.sysfapp_subnet_cidr

    depends_on = [ module.sys_resource_group ]
}


module "ready_azure_functionappiso_pep" {
    source                          = "../../Resources/private-endpoint"
    name                            = "${var.project_name}-${var.environment}-${var.region_short}-sysfappReadyiso-pep"
    location                        = var.region
    resource_group_name             = module.sys_resource_group.name
    subnet_id                       = var.sysfapp_pep_subnet_cidr
    target_resource_id              = module.ready_azure_functioniso_app.function_resource_id
    subresource_names               = [var.const_resource_name_sites]

    depends_on = [ module.ready_azure_function_app ]
} */