variable "aws_region" {
  description = "AWS region for the ECR repositories."
  type        = string
  default     = "us-east-1"
}

variable "name_prefix" {
  description = "Prefix used for all ECR resources."
  type        = string
  default     = "lambda-api-kronk"
}

variable "model_image_tag" {
  description = "Docker tag used for the model image in ECR."
  type        = string
  default     = "latest"
}

variable "lambda_image_tag" {
  description = "Docker tag used for the Lambda runtime image in ECR."
  type        = string
  default     = "latest"
}
