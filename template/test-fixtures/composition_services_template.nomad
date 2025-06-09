[[range $name, $port := . -]]
service {
  name = "global-[[$name]]-check"
  tags = ["global"]
  port = "[[$name]]"
  check {
    name     = "alive"
    type     = "tcp"
    interval = "10s"
    timeout  = "2s"
  }
}
[[- end]]
