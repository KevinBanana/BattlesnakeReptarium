variable "aws_region" {
  description = "AWS Region"

  type    = string
  default = "us-west-2" # This is where battlesnake servers are hosted. Hosting code here for minimal latency.
}

variable "instance_type" {
  description = "AWS Instance Type"

  type    = string
  default = "t2.nano"
}