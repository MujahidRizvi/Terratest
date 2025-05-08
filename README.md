# Azure Infrastructure as Code (IaC) with Terraform & Azure DevOps

This repository manages a **fully private**, **modular**, and **environment-isolated** Azure infrastructure using **Terraform** and **Azure DevOps Pipelines**.

---

## 📦 Directory Structure

```text
.
├── .azure-pipelines/
│   ├── pipelines/                # Environment-specific pipeline definitions (Dev, QA, Prod)
│   └── templates/                # Shared pipeline templates (reusable across environments)
├── Azure/
│   ├── backend.tf                # Terraform remote backend configuration (Azure Storage Account)
│   ├── environments/             # Separate folders for dev, qa, prod with environment-specific Terraform code
│   └── modules/                  # Reusable Terraform modules for common resources (vnet, function app, apim, pep, etc.)
```

---

## 🚀 Features

- Fully **private Azure infrastructure** (no public-facing endpoints)
- **Modular Terraform codebase**: Easily extendable and reusable across environments
- **Isolated environments**: Dev, QA, and Prod are fully separated
- **Dynamic resource linking**: DNS Zones, VNets, Private Endpoints
- **Separate Log Analytics Workspace per environment**
- **Each resource integrates with its respective Log Analytics Workspace** if it supports diagnostics (Function Apps, App Service, AKS, APIM, etc.)
- **Terraform pipelines in Azure DevOps** with manual promotions between environments
- **Support for incremental resource rollout** without impacting existing infrastructure
- **Tagging enforced**: `Name`, `Project: API Ecosystem`

---

## 📂 Resources Deployed Per Environment

Each environment includes:

```text
| Resource                          | Deployment Details |
|:----------------------------------|:-------------------|
| Virtual Network (VNet)           | Hub-Spoke Model with Peering |
| Subnets                          | AppGateway Subnet, FunctionApp Subnet, Private Endpoint Subnet, AKS Subnet, Bastion Subnet |
| Private Azure Application Gateway | Standard_v2 SKU, Private Frontend IP |
| Azure Function Apps              | Windows OS, .NET 8, Private-Only Access with PEP |
| Azure Logic App (Standard SKU)   | Integrated with VNet |
| Azure API Management (APIM)      | Premium SKU, Internal Mode (VNet Injection) |
| Azure Kubernetes Service (AKS)   | Private Cluster, Azure CNI Networking |
| Azure Storage Account            | Private Endpoints for Blob, File, Queue, Table |
| Bastion Host (Jump Server)       | Deployed into Bastion Subnet with limited NSG rules |
| Log Analytics Workspace          | One per environment + Resource-specific logging integration |
```

---

## 🔐 Security by Default

- No public IP addresses on core resources
- All resources accessed via **Private Endpoints**
- **NSG Rules** locked down to VNet or subnet-level
- Log Analytics used for central logging and diagnostics
- Diagnostic Settings auto-attached to Function Apps, App Services, Storage, APIM, etc.

---

## 🧪 Testing Strategy

**Future Plan**:

- Automated **Terratest** pipelines will be integrated into the `.azure-pipelines/` structure.
- Infrastructure tests will include:
  - Peering validation
  - PEP resolution validation
  - Subnet delegation checks
  - Log Analytics diagnostic settings checks
  - Storage Account access restrictions
  - Function App private endpoint DNS resolution

> 🧪 Terratest pipelines are **planned but not yet live**. A separate `terratest/` directory will be introduced soon.

---

## 🌱 Branch Strategy

We use a **feature branch strategy** for safer deployments:

### 🔄 Feature Branch Flow

1. Create a feature branch from `main`:

   ```bash
   git checkout -b feature/add-function-app
   ```

2. Modify or add Terraform code inside `Azure/environments/dev/`.

3. **Do not trigger automatic pipelines** on each push.

4. Manually run the Dev pipeline for testing:
   - `.azure-pipelines/pipelines/dev-deploy.yml`

5. After successful validation:
   - Merge to `main`
   - Promote manually to QA and Prod via their pipelines.

6. Result: **Stable infrastructure**, **zero downtime**, and **controlled changes**.

---

## 🛡️ DNS Strategy

- Hub VNet links all **Private DNS Zones** for core services (Storage, Functions, App Services, APIM, Redis, etc.).
- Spoke VNets link to same Private DNS Zones but **auto-registration is disabled** (except `azure-api.net` for APIM).
- For each Private Endpoint created:
  - **DNS Zone Group** is automatically attached during Terraform run.
  - Auto-registration is managed through explicit `private_dns_zone_group`, **NOT by enabling registration** on VNet links.
- Function Apps, App Services, APIMs, etc., create **A records** dynamically in corresponding DNS Zones during PEP creation.

---

## 📌 Prerequisites

| Tool               | Required Version         |
|:------------------|:--------------------------|
| Terraform          | >= 1.4                   |
| AzureRM Provider   | = 4.25.0                 |
| Azure CLI          | Latest                   |
| Azure DevOps Agent | Hosted or Self-Hosted    |

> Ensure that Service Connections are already configured for pipeline execution.

---

## 📬 Questions?

Open an Issue or reach out to the DevOps / Cloud Engineering team in case of questions, improvements, or incidents related to infrastructure.

---

## ✅ Summary

- Modular, extensible Azure Infra-as-Code
- Private Networking at every layer
- Centralized Logging per Resource
- Future-proofed Testing via Terratest
- GitOps-style Controlled Releases via Azure DevOps Pipelines

---

## 🚀 Let’s Build the Cloud the Right Way! 🚀
