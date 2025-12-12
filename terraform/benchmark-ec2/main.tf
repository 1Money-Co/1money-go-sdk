terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "onemoney-sdk"
      Component = "benchmark-ec2"
      ManagedBy = "terraform"
    }
  }
}

# =============================================================================
# Variables
# =============================================================================

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-2"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "key_name" {
  description = "Name of the SSH key pair (must exist in AWS)"
  type        = string
}

variable "private_key_path" {
  description = "Path to private key file for SSH provisioner"
  type        = string
  default     = "~/.ssh/id_ed25519"
}

# =============================================================================
# Data Sources
# =============================================================================

# Latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

# =============================================================================
# VPC
# =============================================================================

resource "aws_vpc" "benchmark" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "benchmark-vpc"
  }
}

resource "aws_internet_gateway" "benchmark" {
  vpc_id = aws_vpc.benchmark.id

  tags = {
    Name = "benchmark-igw"
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.benchmark.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name = "benchmark-public-subnet"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.benchmark.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.benchmark.id
  }

  tags = {
    Name = "benchmark-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# =============================================================================
# Security Group
# =============================================================================

resource "aws_security_group" "benchmark" {
  name        = "benchmark-ec2-sg"
  description = "Security group for benchmark EC2 instance"
  vpc_id      = aws_vpc.benchmark.id

  # SSH access
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP for testing
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS for API calls
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# =============================================================================
# EC2 Instance
# =============================================================================

resource "aws_instance" "benchmark" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = var.instance_type
  key_name      = var.key_name
  subnet_id     = aws_subnet.public.id

  vpc_security_group_ids = [aws_security_group.benchmark.id]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = {
    Name = "benchmark-ec2"
  }

  # Upload CLI binary
  provisioner "file" {
    source      = "${path.module}/../../bin/onemoney-cli-linux-amd64"
    destination = "/tmp/onemoney-cli"

    connection {
      type        = "ssh"
      user        = "ec2-user"
      private_key = file(pathexpand(var.private_key_path))
      host        = self.public_ip
    }
  }

  # Install CLI and setup environment
  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/onemoney-cli /usr/local/bin/onemoney-cli",
      "sudo chmod +x /usr/local/bin/onemoney-cli",
      "echo 'CLI installed successfully'",
      "onemoney-cli version"
    ]

    connection {
      type        = "ssh"
      user        = "ec2-user"
      private_key = file(pathexpand(var.private_key_path))
      host        = self.public_ip
    }
  }
}

# =============================================================================
# Outputs
# =============================================================================

output "instance_id" {
  description = "EC2 instance ID"
  value       = aws_instance.benchmark.id
}

output "public_ip" {
  description = "Public IP address"
  value       = aws_instance.benchmark.public_ip
}

output "public_dns" {
  description = "Public DNS name"
  value       = aws_instance.benchmark.public_dns
}

output "ssh_command" {
  description = "SSH command to connect"
  value       = "ssh -i ~/.ssh/${var.key_name}.pem ec2-user@${aws_instance.benchmark.public_ip}"
}
