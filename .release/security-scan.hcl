# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

container {
  dependencies    = true
  alpine_security = true

  secrets {
    all = true
  }
}

binary {
  go_modules = true
  osv        = true
  go_stdlib  = true
  oss_index  = false
  nvd        = false

  secrets {
    all = true
  }

  # Triage items that are _safe_ to ignore here. Note that this list should be
  # periodically cleaned up to remove items that are no longer found by the scanner.
  triage {
    suppress {
      vulnerabilities = [
        "GHSA-rx97-6c62-55mf", // https://github.com/github/advisory-database/pull/5759 TODO(dduzgun): remove when dep updated.
      ]
    }
  }
}
