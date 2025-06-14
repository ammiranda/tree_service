terraform {
  backend "s3" {
    bucket         = "tree-service-terraform-state"
    key            = "terraform.tfstate"
    region         = "us-east-2"
    encrypt        = true
    dynamodb_table = "tree-service-terraform-locks"
  }
} 