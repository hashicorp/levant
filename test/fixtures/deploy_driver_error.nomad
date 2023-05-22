# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# tests driver error with an invalid docker image tag

job "[[.job_name]]" {
  datacenters = ["dc1"]
  type        = "service"
  update {
    max_parallel      = 1
    min_healthy_time  = "10s"
    healthy_deadline  = "15s"
    progress_deadline = "20s"
  }

  group "test" {
    count = 1
    restart {
      attempts = 1
      interval = "10s"
      delay    = "5s"
      mode     = "fail"
    }
    ephemeral_disk {
      size = 300
    }
    task "alpine" {
      driver = "docker"
      config {
        image = "alpine:badimagetag"
      }
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
