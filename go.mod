module github.com/hashicorp/levant

go 1.13

require (
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/hashicorp/consul/api v1.8.1
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/nomad v1.1.0
	github.com/hashicorp/nomad/api v0.0.0-20210527173017-41a43a98dc82
	github.com/hashicorp/terraform v0.13.5
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/cli v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.6.0
	github.com/sean-/conswriter v0.0.0-20180208195008-f5ae3917a627
	github.com/sean-/pager v0.0.0-20180208200047-666be9bf53b5 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/hashicorp/nomad/api => github.com/hashicorp/nomad/vendor/github.com/hashicorp/nomad/api v0.0.0-20210517202321-f99f1e27bb66
