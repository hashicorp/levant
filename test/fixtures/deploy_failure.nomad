job "[[.job_name]]" {
  datacenters = ["dc1"]
  type = "service"
  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "15s"
    progress_deadline = "20s"
  }

  group "cache" {
    count = 1
    restart {
      attempts = 1
      interval = "10s"
      delay = "5s"
      mode = "fail"
    }
    ephemeral_disk {
      size = 300
    }
    task "redis" {
      driver = "docker"
      config {
        image = "redis:badimagetag"
      }
      resources {
        cpu    = 100
        memory = 128
        network {
          mbits = 10
        }
      }
    }
  }
}
