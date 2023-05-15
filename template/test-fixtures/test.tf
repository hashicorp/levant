# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "job_name" {
  default = "levantExample"
}

variable "task_resource_cpu" {
  description = "the CPU for the task"
  type        = number
  default     = 1313
}
