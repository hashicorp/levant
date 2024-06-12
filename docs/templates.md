## Templates

Alongside enhanced deployments of Nomad jobs; Levant provides templating functionality allowing for greater flexibility throughout your Nomad jobs files. It also allows the same job file to be used across each environment you have, meaning your operation maturity is kept high. 

### Template Substitution

Levant currently supports `.json`, `.tf`, `.yaml`, and `.yml` file extensions for the declaration of template variables and uses opening and closing double squared brackets `[[ ]]` within the templated job file. This is to ensure there is no clash with existing Nomad interpolation which uses the standard `{{ }}` notation.

#### JSON

JSON as well as YML provide the most flexible variable file format. It allows for descriptive and well organised jobs and variables file as shown below.

Example job template:
```hcl
resources {
    cpu    = [[.resources.cpu]]
    memory = [[.resources.memory]]

    network {
        mbits = [[.resources.network.mbits]]
    }
}
```

Example variable file:
```json
{
    "resources":{
        "cpu":250,
        "memory":512,
        "network":{
            "mbits":10
        }
    }
}
```

#### Terraform

Terraform (.tf) is probably the most inflexible of the variable file formats but does provide an easy to follow, descriptive manner in which to work. It may also be advantageous to use this format if you use Terraform for infrastructure as code thus allow you to use a consistant file format.

Example job template:
```hcl
resources {
    cpu    = [[.resources_cpu]]
    memory = [[.resources_memory]]

    network {
        mbits = [[.resources_network_mbits]]
    }
}
```

Example variable file:
```hcl
variable "resources_cpu" {
  description = "the CPU in MHz to allocate to the task group"
  type        = "string"
  default     = 250
}

variable "resources_memory" {
  description = "the memory in MB to allocate to the task group"
  type        = "string"
  default     = 512
}

variable "resources_network_mbits" {
  description = "the network bandwidth in MBits to allocate"
  type        = "string"
  default     = 10
}
```

#### YAML

Example job template:
```hcl
resources {
    cpu    = [[.resources.cpu]]
    memory = [[.resources.memory]]

    network {
        mbits = [[.resources.network.mbits]]
    }
}
```

Example variable file:
```yaml
---
resources:
  cpu: 250
  memory: 512
  network:
    mbits: 10
```

### Template Functions

Levant's template rendering supports a number of functions which provide flexibility when deploying jobs. As with the variable substitution, it uses opening and closing double squared brackets `[[ ]]` as not to conflict with Nomad's templating standard. Levant parses job files using the [Go Template library](https://golang.org/pkg/text/template/) which makes available the features of that library as well as the functions described below.

If you require any additional functions please raise a feature request against the project.

#### consulKey

Query Consul for the value at the given key path and render the template with the value. In the below example the value at the Consul KV path `service/config/cpu` would be `250`.

Example:
```
[[ consulKey "service/config/cpu" ]]
```

Render:
```
250
```

#### consulKeyExists

Query Consul for the value at the given key path. If the key exists, this will return true, false otherwise. This is helpful in controlling the template flow, and adding conditional logic into particular sections of the job file. In this example we could try and control where the job is configured with a particular "alerting" setup by checking for the existance of a KV in Consul.

Example:
```
{{ if consulKeyExists "service/config/alerting" }}
  <configure alerts>
{{ else }}
  <skip configure alerts>
{{ end }}
```

#### consulKeyOrDefault

Query Consul for the value at the given key path. If the key does not exist, the default value will be used instead. In the following example we query the Consul KV path `service/config/database-addr` but there is nothing at that location. If a value did exist at the path, the rendered value would be the KV value. This can be helpful when configuring jobs which defaults which are appropriate for local testing and development.

Example:
```
[[ consulKeyOrDefault "service/config/database-addr" "localhost:3306" ]]
```

Render:
```
localhost:3306
```

#### env

Returns the value of the given environment variable.

Example:
```
[[ env "HOME" ]]
[[ or (env "NON_EXISTENT") "foo" ]]
```

Render:
```
/bin/bash
foo
```


#### fileContents

Reads the entire contents of the specified file and adds it to the template.

Example file contents:
```
---
yaml:
  - is: everywhere
```

Example job template:
```
[[ fileContents "/etc/myapp/config" ]]
```

Render:
```
---
yaml:
  - is: everywhere
```


#### loop

Accepts varying parameters and differs its behavior based on those parameters as detailed below.

If loop is given a single int input, it will loop up to, but not including the given integer from index 0:

Example:
```
[[ range $i := loop 3 ]]
this-is-loop[[ $i ]][[ end ]]
```

Render:
```
this-is-output0
this-is-output1
this-is-output2
```

If given two integers, this function will begin at the first integer and loop up to but not including the second integer:

Example:
```
[[ range $i := loop 3 6 ]]
this-is-loop[[ $i ]][[ end ]]
```

Render:
```
this-is-output3
this-is-output4
this-is-output5
```

#### parseBool

Takes the given string and parses it as a boolean value which can be helpful in performing conditional checks. In the below example if the key has a value of "true" we could use it to alter what tags are added to the job:

Example:
```
[[ if "true" | parseBool ]][[ "beta-release" ]][[ end ]]
```

Render:
```
beta-release
```

#### parseFloat

Takes the given string and parses it as a base-10 float64.

Example:
```
[[ "3.14159265359" | parseFloat ]]
```

Render:
```
3.14159265359
```

#### parseInt

Takes the given string and parses it as a base-10 int64 and is typically combined with other helpers such as loop:

Example:
```
[[ with $i := consulKey "service/config/conn_pool" | parseInt ]][[ range $d := loop $i ]]
conn-pool-id-[[ $d ]][[ end ]][[ end ]]
```

Render:
```
conn-pool-id-0
conn-pool-id-1
conn-pool-id-2
```

#### parseJSON

Takes the given input and parses the result as JSON. This can allow you to wrap an entire job template as shown below and pull variables from Consul KV for template rendering. The below example is based on the template substitution above and expects the Consul KV to be `{"resources":{"cpu":250,"memory":512,"network":{"mbits":10}}}`:

Example:
```
[[ with $data := consulKey "service/config/variables" | parseJSON ]]
resources {
    cpu    = [[.resources.cpu]]
    memory = [[.resources.memory]]

    network {
        mbits = [[.resources.network.mbits]]
    }
}
[[ end ]]
```

Render:
```
resources {
    cpu    = 250
    memory = 512

    network {
        mbits = 10
    }
}
```

#### parseUint

Takes the given string and parses it as a base-10 int64.

Example:
```
[[ "100" | parseUint ]]
```

Render:
```
100
```

#### replace

Replaces all occurrences of the search string with the replacement string.

Example:
```
[[ "Batman and Robin" | replace "Robin" "Catwoman" ]]
```

Render:
```
Batman and Catwoman
```

#### timeNow

Returns the current ISO_8601 standard timestamp as a string in the timezone of the machine the rendering was triggered on.

Example:
```
[[ timeNow ]]
```

Render:
```
2018-06-25T09:45:08+02:00
```

#### timeNowUTC

Returns the current ISO_8601 standard timestamp as a string in UTC.

Example:
```
[[ timeNowUTC ]]
```

Render:
```
2018-06-25T07:45:08Z
```

#### timeNowTimezone

Returns the current ISO_8601 standard timestamp as a string of the timezone specified. The timezone must be specified according to the entries in the IANA Time Zone database, such as "ASIA/SEOUL". Details of the entries can be found on [wikipedia](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) or your local workstation (Mac or BSD) by searching within `/usr/share/zoneinfo`.

Example:
```
[[ timeNowTimezone "ASIA/SEOUL" ]]
```

Render:
```
2018-06-25T16:45:08+09:00
```

#### toLower

Takes the argument as a string and converts it to lowercase.

Example:
```
[[ "QUEUE-NAME" | toLower ]]
```

Render:
```
queue-name
```

#### toUpper

Takes the argument as a string and converts it to uppercase.

Example:
```
[[ "queue-name" | toUpper ]]
```

Render:
```
QUEUE-NAME
```

#### add

Returns the sum of the two passed values.

Examples:
```
[[ add 5 2 ]]
```

Render:
```
7
```

#### subtract

Returns the difference of the second value from the first.

Example:
```
[[ subtract 2 5 ]]
```

Render:
```
3
```

#### multiply

Returns the product of the two values.

Example:
```
[[ multiply 4 4 ]]
```

Render:
```
16
```

#### divide

Returns the division of the second value from the first.

Example:
```
[[ divide 2 6 ]]
```

Render:
```
3
```

#### modulo

Returns the modulo of the second value from the first.

Example:
```
[[ modulo 2 5 ]]
```

Render:
```
1
```

#### Access Variable Globally

Example config file:
```yaml
my_i32: 1
my_array:
  - "a"
  - "b"
  - "c"
my_nested:
  my_data1: "lorempium"
  my_data2: "faker"
```

Template:
```
[[ $.my_i32 ]]
[[ range $c := $.my_array ]][[ $c ]]-[[ $.my_i32 ]],[[ end ]]
```

Render:
```
1
a1,b1,c1,
```

#### Sprig Template
More about Sprig here: [Sprig](https://masterminds.github.io/sprig/)

#### Sprig Join String

Template: 
```
[[ $.my_array | sprigJoin `-` ]]
```

Render:
```
a-b-c
```

#### Define variable

Template:
```
[[ with $data := $.my_nested ]]
ENV_x1=[[ $data.my_data1 ]]
ENV_x2=[[ $data.my_data2 ]]
[[ end ]] 
```

Render:
```

ENV_x1=lorempium
ENV_x2=faker

```
