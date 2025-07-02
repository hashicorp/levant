// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

job "[[.job_name]]" {
  datacenters = ["dc1"]
  
  task "test" {
    driver = "exec"
    
    config {
      command = "echo"
      args = [
        "DB Host: [[.config.database.host]]",
        "DB Port: [[.config.database.port]]",
        "DB User: [[.config.database.username]]",
        "DB Pass: [[.config.database.password]]",
        "Cache Enabled: [[.config.cache.enabled]]",
        "Cache TTL: [[.config.cache.ttl]]",
        "Log Level: [[.config.logging.level]]"
      ]
    }
  }
}