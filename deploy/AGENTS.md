# AGENTS.md — deploy/

Terraform configuration that deploys the BackRetroLog app on a GCE e2-micro VM
within the GCP free tier.

## Commands

```bash
# Initialize
terraform init

# Validate
terraform validate

# Format
terraform fmt

# Plan
terraform plan -out terraform.plan -var-file=terraform.tfvars

# Apply
terraform apply terraform.plan

# Destroy
terraform destroy
```

## File Map

| File | Purpose |
|------|---------|
| `01-provider.tf` | Google provider config |
| `02-vpc_ip.tf` | Static external IP |
| `03-vm.tf` | COS image, data disk, VM with startup script |
| `04-firewall.tf` | Firewall rule for TCP 80 |
| `05-outputs.tf` | `external_ip`, `app_url` |
| `variables.tf` | All variable declarations with validation |
| `startup.sh.tftpl` | Startup script: format disk, pull image, run container |
| `terraform.tfvars.example` | Template for user variables |

## Variables

- `project_id`, `region`, `zone` — GCP location
- `vm_name`, `ip_name` — resource naming
- `machine_type` — defaults to `e2-micro` (free tier)
- `network` — VPC network (default: `default`)
- `disk_size` — data disk GB (default: 10)
- `docker_image` — image to deploy (required)
- `rawg_api_key` — optional, sensitive
- `base_url` — optional, auto-detected from external IP
- `allowed_source_ranges` — firewall CIDRs (default: `0.0.0.0/0`)

## Deployment Flow

1. `terraform init` (one-time)
2. Copy `terraform.tfvars.example` to `terraform.tfvars`, fill in values
3. `terraform plan` / `terraform apply`
4. App is accessible at the `app_url` output ~2 minutes later
5. To update: change `docker_image` tag, `terraform apply` again
6. Check logs: `gcloud compute ssh <vm_name> -- sudo journalctl -u google-startup-scripts.service --no-pager`

## Code Style

- Same conventions as `gce-postgresql-vm`: numbered TF files, snake_case variables,
  validation blocks, templatefile for startup script, `set -e -x` in scripts.
- Sensitive variables marked `sensitive = true`
- Never commit `terraform.tfvars`, `.terraform/`, or `.plan` files
