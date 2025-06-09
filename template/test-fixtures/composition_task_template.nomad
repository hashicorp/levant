task "[[.Name]]" {
  driver = "docker"
  config {
    image = "[[.Image]]"
    port_map = {
      [[range $name, $port := .Services -]]
      [[$name]] = [[$port]][[end]]
    }
  }

  resources {
    cpu    = 500
    memory = [[.Memory]]
  }

  [[include "test-fixtures/composition_services_template.nomad" .Services]]
}
