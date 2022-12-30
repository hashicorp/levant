# Levant

[![Build Status](https://circleci.com/gh/hashicorp/levant.svg?style=svg)](https://circleci.com/gh/hashicorp/levant) [![Discuss](https://img.shields.io/badge/discuss-nomad-00BC7F?style=flat)](https://discuss.hashicorp.com/c/nomad)

Levant is an open source templating and deployment tool for [HashiCorp Nomad][] jobs that provides
realtime feedback and detailed failure messages upon deployment issues.

## Features

- **Realtime Feedback**: Using watchers, Levant provides realtime feedback on Nomad job deployments
  allowing for greater insight and knowledge about application deployments.

- **Advanced Job Status Checking**: Particularly for system and batch jobs, Levant ensures the job,
  evaluations and allocations all reach the desired state providing feedback at every stage.

- **Dynamic Job Group Counts**: If the Nomad job is currently running on the cluster, Levant dynamically
  updates the rendered template with the relevant job group counts before deployment.

- **Failure Inspection**: Upon a deployment failure, Levant inspects each allocation and logs information
  about each event, providing useful information for debugging without the need for querying the cluster
  retrospectively.

- **Canary Auto Promotion**: In environments with advanced automation and alerting, automatic promotion
  of canary deployments may be desirable after a certain time threshold. Levant allows the user to
  specify a `canary-auto-promote` time period, which if reached with a healthy set of canaries,
  automatically promotes the deployment.

- **Multiple Variable File Formats**: Currently Levant supports `.json`, `.tf`, `.yaml`, and `.yml`
  file extensions for the declaration of template variables.

- **Auto Revert Checking**: In the event that a job deployment does not pass its healthy threshold
  and the job has auto-revert enabled; Levant tracks the resulting rollback deployment so you can
  see the exact outcome of the deployment process.

## Download & Install

- Official Levant binaries can be downloaded from the [HashiCorp releases site][releases-hashicorp].

- Levant can be installed via the go toolkit using `go get github.com/hashicorp/levant && go install github.com/hashicorp/levant`

- A docker image can be found on [Docker Hub][levant-docker]. The latest version can be downloaded
  using `docker pull hashicorp/levant`.

- Levant can be built from source by firstly cloning the repository `git clone git://github.com/hashicorp/levant.git`.
  Once cloned, a binary can be built using the `make dev` command which will be available at
  `./bin/levant`.

- There is a [Levant Ansible role][levant-ansible] available to help installation on machines. Thanks
  to @stevenscg for this.

- Pre-built binaries of Levant from versions 0.2.9 and earlier can be downloaded from the [GitHub releases page][releases] page. These binaries were released prior to the migration to the HashiCorp organization. For example: `curl -L https://github.com/hashicorp/levant/releases/download/0.2.9/linux-amd64-levant -o levant`

## Templating

Levant includes functionality to perform template variables substitution as well as trigger built-in
template functions to add timestamps or retrieve information from Consul. For full details please
consult the [templates][] documentation page.

## Commands

Levant supports a number of command line arguments which provide control over the Levant binary. For
detail about each command and its supported flags, please consult the [commands][] documentation page.

## Clients

Levant utilizes the Nomad and Consul official clients and configuration can be done via a number of
environment variables. For detail about these please read through the [clients][] documentation page.

## Contributing

Community contributions to Levant are encouraged. Please refer to the [contribution guide][] for
details about hacking on Levant.

[clients]: ./docs/clients.md
[commands]: ./docs/commands.md
[templates]: ./docs/templates.md
[contribution guide]: https://github.com/hashicorp/levant/blob/master/.github/CONTRIBUTING.md
[hashicorp nomad]: https://www.nomadproject.io/
[releases]: https://github.com/hashicorp/levant/releases
[levant-docker]: https://hub.docker.com/r/hashicorp/levant/
[levant-ansible]: https://github.com/stevenscg/ansible-role-levant
[releases-hashicorp]: https://releases.hashicorp.com/levant/
