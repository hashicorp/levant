job "[[.job_name]]" {
  datacenters = ["dc1"]
  type = "service"
  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "1m"
    auto_revert      = true
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
    network {
      mode = "bridge"
    }
    service {
      name = "global-redis-check"
      tags = ["global", "cache"]
      port = "6379"

      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "foobar"
              local_bind_port  = 9200
              datacenter       = "[[ .upstream_datacenter ]]"
            }
          }
        }
      }

      check {
        name     = "alive"
        type     = "tcp"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "redis" {
      template {
        data = <<EOH
        APP_ENV={{ key "config/app/env" }}
        APP_DEBUG={{ key "config/app/debug" }}
        APP_KEY={{ secret "secret/key" }}
        APP_URL={{ key "config/app/url" }}
        EOH
        destination = "core/.env"
        change_mode = "noop"
      }

      driver = "docker"
      config {
        image = "redis:3.2"
      }
      resources {
        cpu    = [[.task_resource_cpu]]
        memory = 256
      }
    }
  }
}
