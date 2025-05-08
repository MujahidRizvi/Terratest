variable "project_name" {
    description = "Project name prefix"
    type        = string
}

variable "environment" {
    description = "Environment name (dev, qa, prod)"
    type        = string
}

variable "region" {
    description = "Azure region"
    type        = string
}

variable "region_short" {
    description = "Short version of the Azure region"
    type        = string
}
