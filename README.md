# Levant

[![Build Status](https://travis-ci.org/jrasell/levant.svg?branch=master)](https://travis-ci.org/jrasell/levant) [![Go Report Card](https://goreportcard.com/badge/github.com/jrasell/levant)](https://goreportcard.com/report/github.com/jrasell/levant) [![GoDoc](https://godoc.org/github.com/jrasell/levant?status.svg)](https://godoc.org/github.com/jrasell/levant)

Levant is an open source templating and deployment tool for [HashiCorp Nomad](https://www.nomadproject.io/) jobs that provides realtime feedback and detailed failure messages upon deployment issues.

## Features

* **Realtime Feedback**: Using watchers, Levant provides realtime feedback on Nomad job deployments allowing for greater insight and knowledge about application deployments.

* **Advanced Job Status Checking**: Particulary for system and batch jobs, Levant will ensure the job, evaluations and allocations all reach the desired state providing feedback at every stage.

* **Dynamic Job Group Counts**: If the Nomad job is currently running on the cluster, Levant will dynamically update the rendered template with the relevant job group counts before deployment.

* **Failure Inspection**: Upon a deployment failure, Levant will inspect each allocation and log information about each event, providing useful information for debugging without the need for querying the cluster retrospectively.

* **Canary Auto Promotion**: In environments with advanced automation and alerting, automatic promotion of canary deployments may be desirable after a certain time threshold. Levant allows the user to specify a `canary-auto-promote` time period, which if reached with a healthy set of canaries, will automatically promote the deployment.

* **Multiple Variable File Formats**: Currently Levant supports `.tf`, `.yaml` and `.yml` file extensions for the declaration of template variables. *This is planned to increase in the near future.*

* **Auto Revert Checking**: In the event that a job deployment does not pass its healthy threshold and the job has auto-revert enabled; Levant will track the resulting rollback deployment so you can see the exact outcome of the deployment process.

## Download

* The Levant binary can be downloaded from the [GitHub releases page](https://github.com/jrasell/levant/releases) using `curl -L https://github.com/jrasell/levant/releases/download/0.1.0/linux-amd64-levant -o levant`

* A docker image can be found on [Docker Hub](hub.docker.com/jrasell/levant), the latest version can be downloaded using `docker pull jrasell/levant`.

* Levant can be built from source by firstly cloning the repository `git clone github.com/jrasell/levant.git`. Once cloned the binary can be built using the `make` command or invoking the `build.sh` script located in the scripts directory.

## Variable File Examples

Levant currently supports `.tf`, `.yaml` and `.yml` file extensions for the declaration of template variables and uses opening and closing double squared brackets `[[ ]]` within the templated job file. This is to ensure there is no clash with existing Nomad interpolation which uses the standard `{{ }}` notation.

Job Template:
```hcl
job "[[.job_name]]" {
  datacenters = ["dc1"]
  ...
```

`.tf` variables file:
```hcl
variable "job_name" {
  default = "jrasell-example"
}
```

`.yaml` or `.yml` variables file:
```yaml
job_name: jrasell-example
```

## Commands

Levant supports a number of command line arguments which provide control over the Levant binary. Levant also supports autoloading files; where if `levant [deploy,render]` is run, Levant will look in the current working directory for a `levant.[yaml,yml,tf]` file and a single `*.nomad` file to use for the command actions.

### Command: `deploy`

`deploy` is the main entry point into Levant for deploying a Nomad job and supports the following flags which should then be proceeded by the Nomad job template you which to deploy. An example deployment command would look like `levant deploy -log-level=debug example.nomad`.

* **-address** (string: "http://localhost:4646") The HTTP API endpoint for Nomad where all calls will be made.

* **-canary-auto-promote** (int: 0) The time period in seconds that Levant should wait for before attempting to promote a canary deployment.

* **-force-count** (bool: false) Use the taskgroup count from the Nomad job file instead of the count that is obtained from the running job count.

* **-log-level** (string: "INFO") The level at which Levant will log to. Valid values are DEBUG, INFO, WARNING, ERROR and FATAL.

* **-var-file** (string: "") The variables file to render the template with.

The `deploy` command also supports passing variables individually on the command line. Multiple commands can be passed in the format of `-var 'key=value'`. Variables passed via the command line take precedence over the same variable declared within a passed variable file.

Full example:

```
levant deploy -log-level=debug -address=nomad.devoops -var-file=var.yaml -var 'var=test' example.nomad
```

### Dispatch: `dispatch`

`dispatch` allows you to dispatch an instance of a Nomad parameterized job and utilise Levant's advanced job checking features to ensure the job reaches the correct running state.

* **-address** (string: "http://localhost:4646") The HTTP API endpoint for Nomad where all calls will be made.

* **-log-level** (string: "INFO") The level at which Levant will log to. Valid values are DEBUG, INFO, WARNING, ERROR and FATAL.

* **-meta** (string: "key=vaule") The metadata key will be merged into the job's metadata. The job may define a default value for the key which is overridden when dispatching. The flag can be provided more than once to inject multiple metadata key/value pairs. Arbitrary keys are not allowed. The parameterized job must allow the key to be merged.

The command also supports the ability to send data payload to the dispatched instance. This can be provided via stdin by using "-" for the input source or by specifying a path to a file.

Full example:

```
levant dispatch -log-level=debug -address=nomad.devoops -meta key=value dispatch_job payload_item
```

### Command: `render`

`render` allows rendering of a Nomad job template without deploying, useful when testing or debugging. An example render command would look like `levant render -out job.nomad job.nomad.tpl`, options:

* **-var-file** (string: "") The variables file to render the template with.

* **-output** (string: "") The path to write the rendered template to. The template will be rendered to stdout if this is not set.

Like `deploy`, the `render` command also supports passing variables individually on the command line. Multiple vars can be passed in the format of `-var 'key=value'`. Variables passed via the command line take precedence over the same variable declared within a passed variable file.

Full example:

```
levant render -var-file=var.yaml -var 'var=test' example.nomad
```

### Command: `version`

The `version` command displays build information about the running binary, including the release version.

## Nomad Client

The project uses the Nomad [Default API Client](https://github.com/hashicorp/nomad/blob/master/api/api.go#L191) which means the following Nomad client parameters used by Levant are configurable via environment variables:

 * **NOMAD_ADDR** - The address of the Nomad server.
 * **NOMAD_REGION** - The region of the Nomad servers to forward commands to.
 * **NOMAD_NAMESPACE** - The target namespace for queries and actions bound to a namespace.
 * **NOMAD_CACERT** - Path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate.
 * **NOMAD_CAPATH** - Path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate.
 * **NOMAD_CLIENT_CERT** - Path to a PEM encoded client certificate for TLS authentication to the Nomad server.
 * **NOMAD_CLIENT_KEY** - Path to an unencrypted PEM encoded private key matching the client certificate from `NOMAD_CLIENT_CERT`.
 * **NOMAD_SKIP_VERIFY** - Do not verify TLS certificate.
 * **NOMAD_TOKEN** - The SecretID of an ACL token to use to authenticate API requests with.

## Contributing

Contributions to Levant are very welcome! Please refer to our [contribution guide](https://github.com/jrasell/levant/blob/master/.github/CONTRIBUTING.md) for details about hacking on Levant.
