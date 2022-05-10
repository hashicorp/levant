# TODO: do we need container if skipping docker steps?
container {
  dependencies = true
  alpine_secdb = true
  secrets      = true
}

binary {
  secrets    = true
  go_modules = true
  osv        = true
  oss_index  = false
  nvd        = false
}
