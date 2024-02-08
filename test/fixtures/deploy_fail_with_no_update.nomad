job "[[.job_name]]" {
  datacenters = ["dc1"]
  type = "service"

  group "test" {
    count = 1

    restart {
      attempts = 1
      interval = "5s"
      delay = "1s"
      mode = "fail"
    }

    ephemeral_disk {
      size = 300
    }

    update {
      max_parallel     = 1
      min_healthy_time = "10s"
      healthy_deadline = "1m"
    }

    network {
      port "http" {
        to = 80
      }
    }

    service {
      name = "fake-service"
      port = "http"

       check {
         name     = "alive"
         type     = "tcp"
         interval = "10s"
         timeout  = "2s"
       }
    }

    task "alpine" {
      driver = "docker"
      config {
        image = "alpine"
        command = "sleep 1 && exit 1"
      }
      resources {
        cpu    = 100
        memory = 20
        network {
          mbits = 10
        }
      }
    }
  }
}
