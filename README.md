# Levant

[![Build Status](https://travis-ci.org/jrasell/levant.svg?branch=master)](https://travis-ci.org/jrasell/levant) [![Go Report Card](https://goreportcard.com/badge/github.com/jrasell/levant)](https://goreportcard.com/report/github.com/jrasell/levant) [![GoDoc](https://godoc.org/github.com/jrasell/levant?status.svg)](https://godoc.org/github.com/jrasell/levant)

*CURRENTLY UNDER DEVELOPMENT*

Levant is an open source templating and deployment tool for [HashiCorp Nomad](https://www.nomadproject.io/) jobs that provides realtime feedback and detailed failure messages upon deployment issues.

## Features

* **Realtime Feedback**: Using watchers, Levant provides realtime feedback on Nomad job deployments allowing for greater insight and knowledge about application deployments.

* **Dynamic Job Group Counts**: If the Nomad job is currently running on the cluster, Levant will dynamically update the rendered template with the relevant job group counts before deployment.

* **Failure Inspection**: Upon a deployment failure, Levant will inspect each allocation and log information about each event, providing useful information for debugging without the need for querying the cluster retrospectively.

* **Multiple Variable File Formats**: Currently Levant supports `.tf`, `.yaml` and `.yml` file extensions for the declaration of template variables. *This is planned to increase in the near future.*

## Download

* The Levant binary can be downloaded from the [GitHub releases page]() using `curl https://github.com/jrasell/levant/releases/download/v0.0.1/linux-amd64-levant -o levant`

* A docker image can be found on [Docker Hub](hub.docker.com/jrasell/levant), the latest version can be downloaded using `docker pull jrasell/levant`.

* Levant can be built from source by firstly cloning the repository `git clone github.com/jrasell/levant.git`. Once cloned the binary can be built using the `make` command or invoking the `build.sh` script located in the scripts director.

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

Levant supports a number of command line arguments which provide control over the Levant binary.

### Command: `deploy`

`Deploy` is the main entry point into Levant for deploying a Nomad job and supports the following flags which should then be proceded by the Nomad job template you which to deploy. An example deployment command would look like `levant -log-level=debug example.nomad`.

* **-address** (string: "http://localhost:4646") The HTTP API endpoint for Nomad where all calls will be made.

* **-log-level** (string: "INFO") The level at which Levant will log to. Valid values are DEBUG, INFO, WARNING, ERROR and FATAL.

* **-var-file** (string: "") The variables file to render the template with.

The `deploy` command also supports passing variables individually on the command line. Multiple commands can be passed in the format of `-var 'key=value'`. Variables passed via the command line take precedence over the same variable declared within a passed variable file.

Full example:

```
levant deploy -log-level=debug -address=nomad.devoops -var-file=var.yaml -var 'var=test' example.nomad
```

### Command: `render`

`render` allows rendering of a Nomad job template without deploying, useful when testing or debugging. An example render command would look like `levant render -out job.nomad job.nomad.tpl`, options:

* **-var-file** (string: "") The variables file to render the template with.

* **-output** (string: "") The path to write the rendered template to. The template will be rendered to stdout if this is not set.

Like `deploy`, the `render` command also supports passing variables individually on the command line. Multiple vars can be passed in the format of `-var 'key=value'`. Variables passed via the command line take precedence over the same variable declared within a passed variable file.

Full example:

```
levant render -var-file=var.yaml -var 'var=test' render example.nomad
```

### Command: `version`

The `version` command displays build information about the running binary, including the release version.
