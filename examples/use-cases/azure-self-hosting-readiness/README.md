# Azure self-hosting readiness

This example prepares an Azure substrate for a future E2B BYOC or self-hosted deployment. It does not claim that E2B Azure BYOC is generally available today.

E2B's hosted BYOC documentation currently says BYOC is available for AWS and GCP and that Azure support is in progress. The open-source `e2b-dev/infra` repository lists Azure in its README, but its self-hosting guide currently documents GCP and AWS steps, not an Azure deployment path.

The example creates the Azure resources that an Azure E2B deployment would need first:

- Resource group
- Virtual network and workload subnets
- Network security group
- Storage account for templates, snapshots, and logs
- Azure Container Registry for template images
- Key Vault for operator-managed secrets
- Log Analytics workspace
- User-assigned managed identity for an orchestrator or provisioning agent

Validate without creating resources:

```shell
terraform init -backend=false
terraform validate
terraform plan
```

Apply only after E2B publishes an Azure BYOC/self-hosting path or gives you onboarding-specific requirements.
