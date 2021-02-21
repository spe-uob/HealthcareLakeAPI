terraform {
  backend "remote" {
    organization = "healthcare-lake"

    workspaces {
      name = "HealthcareDataLakeAPI"
    }
  }
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}
provider "aws" {
  region = var.region
}

module "lambda" {
  source              = "./modules/lambda"
  dynamodb_table_name = module.dynamodb.table_name
  dynamodb_arn        = module.dynamodb.arn
  kms_arn             = module.dynamodb.kms_arn
}

module "dynamodb" {
  source = "./modules/dynamodb"
  stage  = var.stage
}

module "api_gateway" {
  source            = "./modules/api_gateway"
  stage             = var.stage
  lambda_invoke_arn = module.lambda.invoke_arn
  lambda_name       = module.lambda.function_name

  depends_on = [
    module.lambda
  ]
}