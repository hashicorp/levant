## Clients

Levant uses Nomad and Consul clients in order to perform its work. Currently only the HTTP address client parameter can be configured for each client via CLI flags; a choice made to keep the number of flags low. In order to further configure the clients you can use environment variables as detailed below.

### Nomad Client

The project uses the Nomad [Default API Client](https://github.com/hashicorp/nomad/blob/master/api/api.go#L201) which means the following Nomad client parameters used by Levant are configurable via environment variables:

 * **NOMAD_ADDR** - The address of the Nomad server.
 * **NOMAD_REGION** - The region of the Nomad servers to forward commands to.
 * **NOMAD_NAMESPACE** - The target namespace for queries and actions bound to a namespace.
 * **NOMAD_CACERT** - Path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate.
 * **NOMAD_CAPATH** - Path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate.
 * **NOMAD_CLIENT_CERT** - Path to a PEM encoded client certificate for TLS authentication to the Nomad server.
 * **NOMAD_CLIENT_KEY** - Path to an unencrypted PEM encoded private key matching the client certificate from `NOMAD_CLIENT_CERT`.
 * **NOMAD_SKIP_VERIFY** - Do not verify TLS certificate.
 * **NOMAD_TOKEN** - The SecretID of an ACL token to use to authenticate API requests with.

### Consul Client

The project also uses the Consul [Default API Client](https://github.com/hashicorp/consul/blob/master/api/api.go#L282) which means the following Consul client parameters used by Levant are configurable via environment variables:

 * **CONSUL_CACERT** - Path to a CA file to use for TLS when communicating with Consul.
 * **CONSUL_CAPATH** - Path to a directory of CA certificates to use for TLS when communicating with Consul.
 * **CONSUL_CLIENT_CERT** - Path to a client cert file to use for TLS when 'verify_incoming' is enabled.
 * **CONSUL_CLIENT_KEY** - Path to a client key file to use for TLS when 'verify_incoming' is enabled.
 * **CONSUL_HTTP_ADDR** - The `address` and port of the Consul HTTP agent. The value can be an IP address or DNS address, but it must also include the port.
 * **CONSUL_TLS_SERVER_NAME** - The server name to use as the SNI host when connecting via TLS.
 * **CONSUL_HTTP_TOKEN** - ACL token to use in the request. If unspecified, the query will default to the token of the Consul agent at the HTTP address.
