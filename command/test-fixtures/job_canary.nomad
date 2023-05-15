# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

job "example" {
  datacenters = ["dc1"]
  type = "service"
  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "1m"
    auto_revert      = true
    canary           = 1
  }
  group "cache" {
    count = 1
    restart {
      attempts = 10
      interval = "5m"
      delay = "25s"
      mode = "delay"
    }
    ephemeral_disk {
      size = 300
    }
    task "redis" {
      artifact {
        source = "google.com"
      }

      driver = "docker"
      config {
        image = "redis:3.2"
        port_map {
          db = 6379
        }
      }
      resources {
        cpu    = 500
        memory = 256
        network {
          mbits = 10
          port "db" {}
        }
      }
      service {
        name = "global-redis-check"
        tags = ["global", "cache"]
        port = "db"
        check {
          name     = "alive"
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}
