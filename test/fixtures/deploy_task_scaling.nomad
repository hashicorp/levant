# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

job "[[.job_name]]" {
  datacenters = ["dc1"]
  type        = "service"
  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "1m"
    auto_revert      = true
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
      resources {
        cpu    = 100
        memory = 128
      }
      scaling "cpu" {
        policy {
          check "95pct" {
            strategy "app-sizing-percentile" {
              percentile = "95"
            }
          }
        }
      }
      scaling "mem" {
        policy {
          check "max" {
            strategy "app-sizing-max" {}
          }
        }
      }
    }
  }
}
