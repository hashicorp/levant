enable_debug = true

log_level = "debug"

data_dir = "/opt/nomad/data"

bind_addr = "0.0.0.0"

# Enable the client
client {
  enabled = true

  options {
    "driver.raw_exec.enable"    = "1"
    "docker.privileged.enabled" = "true"
  }

  meta {
    "rack" = "r1"
  }

  host_volume "shared_data" {
    path = "/tmp/data"
  }
}

consul {
  address = "127.0.0.1:8500"
}

vault {
  enabled = true
  address = "http://active.vault.service.consul:8200"
}

telemetry {
  collection_interval        = "1s"
  disable_hostname           = true
  prometheus_metrics         = true
  publish_allocation_metrics = true
  publish_node_metrics       = true
}
