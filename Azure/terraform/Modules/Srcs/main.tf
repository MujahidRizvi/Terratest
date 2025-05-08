module "sys_resource_group" {
    source                          = "../../Resources/resource-group"
    name                            = "${var.project_name}-${var.environment}-${var.region_short}-srcs-rg" 
    location                        = var.region
}
