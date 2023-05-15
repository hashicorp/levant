# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

job "periodic_batch_test" {
  datacenters = ["dc1"]
  region      = "global"
  type        = "batch"
  priority    = 75
  periodic {
    cron             = "* 1 * * * *"
    prohibit_overlap = true
  }
  group "periodic_batch" {
    task "periodic_batch" {
      driver = "docker"
      config {
        image = "cogniteev/echo"
      }
      resources {
        cpu    = 100
        memory = 128
        network {
          mbits = 1
        }
      }
    }
  }
}
