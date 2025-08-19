job "composedJob" {
  datacenters = ["dc1"]
  type = "service"
  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "1m"
    auto_revert      = true
  }

  group "composedGroup" {
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
[[range $task := .tasks -]]
[[include "test-fixtures/composition_task_template.nomad" $task | indent 2]]
[[end]]
  }
}
