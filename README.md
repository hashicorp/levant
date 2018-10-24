# Levant

[![Build Status](https://travis-ci.org/jrasell/levant.svg?branch=master)](https://travis-ci.org/jrasell/levant) [![Go Report Card](https://goreportcard.com/badge/github.com/jrasell/levant)](https://goreportcard.com/report/github.com/jrasell/levant) [![GoDoc](https://godoc.org/github.com/jrasell/levant?status.svg)](https://godoc.org/github.com/jrasell/levant)
[![Join the chat at https://gitter.im/levantdeployment/Lobby](https://badges.gitter.im/levantdeployment/Lobby/Lobby.svg)](https://gitter.im/levantdeployment/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Levant is an open source templating and deployment tool for [HashiCorp Nomad](https://www.nomadproject.io/) jobs that provides realtime feedback and detailed failure messages upon deployment issues.

## Features

* **Realtime Feedback**: Using watchers, Levant provides realtime feedback on Nomad job deployments allowing for greater insight and knowledge about application deployments.

* **Advanced Job Status Checking**: Particulary for system and batch jobs, Levant will ensure the job, evaluations and allocations all reach the desired state providing feedback at every stage.

* **Dynamic Job Group Counts**: If the Nomad job is currently running on the cluster, Levant will dynamically update the rendered template with the relevant job group counts before deployment.

* **Failure Inspection**: Upon a deployment failure, Levant will inspect each allocation and log information about each event, providing useful information for debugging without the need for querying the cluster retrospectively.

* **Canary Auto Promotion**: In environments with advanced automation and alerting, automatic promotion of canary deployments may be desirable after a certain time threshold. Levant allows the user to specify a `canary-auto-promote` time period, which if reached with a healthy set of canaries, will automatically promote the deployment.

* **Multiple Variable File Formats**: Currently Levant supports `.json`, `.tf`, `.yaml`, and `.yml` file extensions for the declaration of template variables.

* **Auto Revert Checking**: In the event that a job deployment does not pass its healthy threshold and the job has auto-revert enabled; Levant will track the resulting rollback deployment so you can see the exact outcome of the deployment process.

## Download & Install

* The Levant binary can be downloaded from the [GitHub releases page](https://github.com/jrasell/levant/releases) using `curl -L https://github.com/jrasell/levant/releases/download/0.2.4/linux-amd64-levant -o levant`

* A docker image can be found on [Docker Hub](https://hub.docker.com/r/jrasell/levant/), the latest version can be downloaded using `docker pull jrasell/levant`.

* Levant can be built from source by firstly cloning the repository `git clone github.com/jrasell/levant.git`. Once cloned the binary can be built using the `make` command or invoking the `build.sh` script located in the scripts directory.

* There is a [Levant Ansible role](https://github.com/stevenscg/ansible-role-levant) available to help installation on machines. Thanks to @stevenscg for this.

## Templating

Levant includes functionality to perform template variables substituion as well as trigger built-in template function to add timestamps or retrive information from Consul. For full details please check out the [templates](./docs/templates.md) documentation page.

## Commands

Levant supports a number of command line arguments which provide control over the Levant binary. For detail about each commands and its supported flags, please head over to the [commands](./docs/commands.md) documentation page.

## Clients

Levant utilises the Nomad and Consul offical clients and configuration can be done via a number of environment variables. For detail about these please read through the [clients](./docs/clients.md) documentation page.

## Contributing

Contributions to Levant are very welcome! Please refer to our [contribution guide](https://github.com/jrasell/levant/blob/master/.github/CONTRIBUTING.md) for details about hacking on Levant.
