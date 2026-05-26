data "terraform_remote_state" "ecr" {
  backend = "local"

  config = {
    path = var.ecr_state_path
  }
}
