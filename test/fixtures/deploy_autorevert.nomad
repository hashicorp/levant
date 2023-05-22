# tests a autorevert watcher on a job with its
# namespace set in the job specification itself.
#
# a bad check should cause the job to try reverting; the
# test manipulates the .to value to cause the tests to pass
# on initial deployment and fail on a second one.
# https://github.com/hashicorp/levant/issues/426

job "[[.job_name]]" {
  datacenters = ["dc1"]
  type        = "service"
  namespace   = "test"
  update {
    max_parallel      = 1
    min_healthy_time  = "2s"
    healthy_deadline  = "5s"
    auto_revert       = true
    progress_deadline = "10s"
  }

  group "test" {
    count = 1
    network {
      port "ok" {
        to = [[.to]]
      }
    }
    service {
      provider = "nomad"
      name     = "ok"
      port     = "ok"
      check {
        type     = "http"
        path     = "/"
        interval = "1s"
        timeout  = "1s"
      }
    }
    task "alpine" {
      driver = "docker"
      config {
        image   = "alpine/socat"
        args    = ["tcp-listen:1234,fork,reuseaddr", "system:\"echo HTTP/1.1 200 OK; echo Content-Type: text/plain; echo; echo ok;\""]
        ports   = ["ok"]
      }
      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}