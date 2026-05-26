output "ecr_model_repository_url" {
  description = "ECR repository URL for the model image."
  value       = aws_ecr_repository.model.repository_url
}

output "ecr_lambda_repository_url" {
  description = "ECR repository URL for the Lambda runtime image."
  value       = aws_ecr_repository.lambda.repository_url
}

output "docker_login_command" {
  description = "Command to authenticate Docker against ECR."
  value       = "aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com"
}

output "docker_build_model_command" {
  description = "Example command to build and tag the model image."
  value       = "docker build -f models/Dockerfile -t ${aws_ecr_repository.model.repository_url}:${var.model_image_tag} ."
}

output "docker_push_model_command" {
  description = "Example command to push the model image."
  value       = "docker push ${aws_ecr_repository.model.repository_url}:${var.model_image_tag}"
}

output "docker_build_lambda_command" {
  description = "Example command to build and tag the Lambda runtime image."
  value       = "docker build --build-arg MODEL_IMAGE=${aws_ecr_repository.model.repository_url}:${var.model_image_tag} -t ${aws_ecr_repository.lambda.repository_url}:${var.lambda_image_tag} ."
}

output "docker_push_lambda_command" {
  description = "Example command to push the Lambda runtime image."
  value       = "docker push ${aws_ecr_repository.lambda.repository_url}:${var.lambda_image_tag}"
}
