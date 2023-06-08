# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# tests a canary deployment

job "[[.job_name]]" {
  datacenters = ["dc1"]
  type        = "service"
  update {
    max_parallel     = 1
    min_healthy_time = "2s"
    healthy_deadline = "1m"
    auto_revert      = true
    canary           = 1
  }

  group "test" {
    count = 1
    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }
    ephemeral_disk {
      size = 300
    }
    task "alpine" {
      driver = "docker"
      config {
        image   = "alpine"
        command = "tail"
        args    = ["-f", "/dev/null"]
      }
      env {
        version = "[[ .env_version ]]"
      }
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
