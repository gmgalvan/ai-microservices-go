# Lambda API Kronk

AWS Lambda experiment using Kronk, API Gateway, and container images.

Routes:

- `GET /`
- `GET /api/hello`
- `GET /api/bye`

## Project Layout

- `main.go`: Lambda handler
- `Dockerfile`: Lambda runtime image
- `models/Dockerfile`: preloaded model image
- `models/preload/main.go`: helper that downloads `llama.cpp` and the `.gguf` model during image build
- `infra/ecr/`: Terraform layer that creates ECR repositories
- `infra/service/`: Terraform layer that creates Lambda, API Gateway, IAM, and reads the ECR state

## Architecture

This project uses two container images:

1. Model image
   - preloads `llama.cpp`
   - preloads the GGUF model
   - stores everything under `/opt/kronk`

2. Lambda image
   - builds the Go Lambda handler
   - copies `/opt/kronk` from the model image
   - starts the Lambda handler with `KRONK_BASE_PATH=/opt/kronk`

This avoids downloading the model on every request.

## Deployment Flow

The infrastructure is intentionally split into two Terraform layers.

### Layer 1: ECR

This layer creates:

- the ECR repository for the model image
- the ECR repository for the Lambda image

Run:

```bash
cd lambda-api-kronk/infra/ecr
terraform init
terraform apply
```

Useful outputs:

```bash
terraform output ecr_model_repository_url
terraform output ecr_lambda_repository_url
terraform output docker_login_command
terraform output docker_build_model_command
terraform output docker_push_model_command
terraform output docker_build_lambda_command
terraform output docker_push_lambda_command
```

### Build and Push the Images

From the project root:

```bash
cd lambda-api-kronk
```

Authenticate Docker against ECR:

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com
```

Build and push the model image:

```bash
docker build -f models/Dockerfile -t <model-repo-url>:latest .
docker push <model-repo-url>:latest
```

Build and push the Lambda image:

```bash
docker build --build-arg MODEL_IMAGE=<model-repo-url>:latest -t <lambda-repo-url>:latest .
docker push <lambda-repo-url>:latest
```

### Layer 2: Service

This layer creates:

- the Lambda function
- the HTTP API Gateway
- IAM role and permissions
- CloudWatch log group

It reads the ECR outputs through `terraform_remote_state` from:

```bash
../ecr/terraform.tfstate
```

Run:

```bash
cd lambda-api-kronk/infra/service
terraform init
terraform apply
```

## Testing the API

From `infra/service`:

```bash
terraform output hello_url
terraform output bye_url
terraform output hello_curl
terraform output bye_curl
```

Or call the endpoints directly:

```bash
curl "$(terraform output -raw hello_url)"
curl "$(terraform output -raw bye_url)"
```

## Cold Start Behavior

This project avoids model downloads at request time, but it does not avoid model loading.

What is already solved:

- the model is baked into the image
- `llama.cpp` is baked into the image
- Lambda does not need to download the model during a request

What still happens on a cold start:

- the Lambda execution environment starts
- Kronk initializes
- the model is loaded from `/opt/kronk` into memory

That load step can still be expensive enough to cause:

- `503 Service Unavailable` from API Gateway on the first request
- a successful response on the second request, once the Lambda is warm

This is expected for heavy cold-start workloads.

## How to Mitigate Cold Starts

Recommended options:

1. Use Provisioned Concurrency
   - best option if you want the first request to succeed consistently
   - keeps one or more execution environments warm

2. Increase Lambda memory
   - more memory gives more CPU
   - model initialization often becomes noticeably faster

3. Use a smaller model
   - smaller GGUF models reduce startup and load time

4. Add a warm-up strategy
   - periodic invocations can reduce the chance of a cold start
   - this is less reliable than Provisioned Concurrency

5. Avoid synchronous first-request latency requirements
   - if the workload is too heavy, consider asynchronous invocation patterns

## Verifying That Downloads Are No Longer Happening

Watch the Lambda logs:

```bash
aws logs tail /aws/lambda/lambda-api-kronk --follow --region us-east-1
```

With the image-based approach working correctly, you should see:

- `download-libraries: already installed`
- `download-model: already installed`

You should not see repeated full model downloads on every request.

## Direct Lambda Invocation

To distinguish API Gateway timeout behavior from Lambda runtime behavior, invoke the Lambda directly:

```bash
aws lambda invoke \
  --function-name lambda-api-kronk \
  --region us-east-1 \
  --cli-binary-format raw-in-base64-out \
  --payload '{"version":"2.0","rawPath":"/api/hello"}' \
  /tmp/lambda-response.json

cat /tmp/lambda-response.json
```

If direct invocation works but API Gateway returns `503` on the first request, the issue is almost certainly cold-start latency rather than model packaging.
