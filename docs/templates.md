## Templates

Alongside enhanced deployments of Nomad jobs; Levant provides templating functionality allowing for greater flexibility throughout your Nomad jobs files. It also allows the same job file to be used across each environment you have, meaning your operation maturity is kept high. 

### Template Substitution

Levant currently supports `.tf`, `.yaml` and `.yml` file extensions for the declaration of template variables and uses opening and closing double squared brackets `[[ ]]` within the templated job file. This is to ensure there is no clash with existing Nomad interpolation which uses the standard `{{ }}` notation.

Example Job Template:
```hcl
resources {
    cpu    = [[.cpu]]
    memory = [[.memory]]

    network {
        mbits = [[.mbits]]
    }
}
```

`.tf` variables file:
```hcl
variable "cpu" {
  default = 250
}

variable "memory" {
  default = 512
}

variable "mbits" {
  default = 10
}
```

`.yaml` or `.yml` variables file:
```yaml
cpu: 250
memory: 512
mbits: 10
```

Render:
```hcl
resources {
    cpu    = 250
    memory = 512

    network {
        mbits = 10
    }
}
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