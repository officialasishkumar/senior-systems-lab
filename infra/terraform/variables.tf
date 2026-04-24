variable "name" {
  type    = string
  default = "netops-lab"
}

variable "region" {
  type    = string
  default = "us-east-1"
}

variable "cidr_block" {
  type    = string
  default = "10.42.0.0/16"
}

variable "availability_zones" {
  type    = list(string)
  default = ["us-east-1a", "us-east-1b"]
}

variable "public_subnets" {
  type    = list(string)
  default = ["10.42.0.0/24", "10.42.1.0/24"]
}

variable "private_subnets" {
  type    = list(string)
  default = ["10.42.10.0/24", "10.42.11.0/24"]
}

