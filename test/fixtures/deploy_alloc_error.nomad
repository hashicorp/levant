# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# test alloc error with a command failure

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
        image   = "alpine"
        command = "badcommandname"
      }
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
