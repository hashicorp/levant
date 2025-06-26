variable "job_name" {
  default = "levantExample"
}

variable "config" {
  default = {
    database = {
      host = "localhost"
      port = 5432
    }
    cache = {
      enabled = true
      ttl = 300
    }
  }
}