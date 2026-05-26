# Infra Layers

- `ecr/`: first layer, creates the ECR repositories
- `service/`: second layer, creates Lambda + API Gateway and reads the outputs from `ecr/` using `terraform_remote_state`
