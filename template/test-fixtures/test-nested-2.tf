variable "config" {
  default = {
    database = {
      username = "admin"
      password = "secret"
    }
    logging = {
      level = "debug"
    }
  }
}