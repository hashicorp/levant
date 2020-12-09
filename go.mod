module github.com/hashicorp/levant

go 1.13

require (
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/hashicorp/consul/api v1.7.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl/v2 v2.7.1 // indirect
	github.com/hashicorp/nomad v0.12.5-0.20201208211811-bbf1eaa833ed
	github.com/hashicorp/nomad/api v0.0.0-20201208211811-bbf1eaa833ed
	github.com/hashicorp/serf v0.9.5 // indirect
	github.com/hashicorp/terraform v0.13.5
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/cli v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.6.0
	github.com/sean-/conswriter v0.0.0-20180208195008-f5ae3917a627
	github.com/sean-/pager v0.0.0-20180208200047-666be9bf53b5 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/zclconf/go-cty v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/sys v0.0.0-20201029020603-3518587229cd // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/hashicorp/nomad/api => github.com/hashicorp/nomad/vendor/github.com/hashicorp/nomad/api v0.0.0-20201029214628-fa2008a42bae
