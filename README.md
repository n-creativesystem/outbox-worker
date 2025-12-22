# Overview

This library implements an outbox worker which periodically fetches messages from an outbox message table and publishes them to SNS topics.

# Usage

## Outbox message table

Applications MUST create an outbox message table with the following columns:

|Column|Type|Description|Updater|
|-|-|-|-|
|`id`|bigint, PK|ID|app|
|`aggregate_type`|varchar(255), not null|SNS topic name|app|
|`aggregate_id`|varchar(128), not null|Message group ID|app|
|`event`|varchar(255), not null|Event type. This value will be set to SNS message attribute `Event`.|app|
|`payload`|JSON, not null|Message body|app|
|`retry_at`|datetime, null|Next retry time when the outbox worker fails to publish the message to SNS|outbox worker|
|`retry_count`|int, null|The number of retry attempts|outbox worker|

The column names are fixed at the moment, but we may allow applications to define dynamic column names in the future.

Applications should migrate the outbox message table.

Additional columns may be present in the outbox message table to suit specific application requirements. For example:
  - `tenant_uid` or `office_id`: for sharding and data analysis
  - `created_at`: the date and time when the record was created
  - `updated_at`: the date and time when the record was updated
    - Note: It is recommended that this column be updated automatically in the database, as the outbox worker will not update it even if it exists.

## Configurations

The following configurations can be defined in a YAML config file. Refer to [examples/config/outbox_polling.yaml](examples/config/outbox_polling.yaml) for a sample.

|Field|Required|Description|
|-|-|-|
|`outbox`|Yes|Outbox worker configurations|
|`database`|Yes|Database configurations|
|`ssh`|No|SSH configurations|
|`publisher`|Yes|Publisher configurations|
|`logging`|No|Logging configurations|
|`tracking`|No|Tracking configurations|

### Outbox worker configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`schema`|No|String|||Database schema|
|`tableName`|Yes|String|`outbox`|`outbox_messages`|Outbox message table name|
|`producerName`|Yes|String||`NeoAP`|Producer name. This value will be set to SNS message attribute `Producer`.|
|`pollingInterval`|No|Time duration|`5s`|`5s`|The time interval between consecutive polling attempts|
|`retryCount`|No|Number|`10`|`10`|The maximum number of retry attempts that the outbox worker will make for a message when it fails to send the message to SNS.|
|`retryBackoff`|No|Time duration|`20s`|`20s`|The fundamental time duration used in calculating the next retry time by the formula `nextRetryTime = currentTime + retryBackoff * (2 ** currentRetryCount)`.|
|`throughput`|No|Number|`3000`|`3000`|The maximum number of messages to send to an SNS topic per second. Max value: 3000.|
|`findEventLimit`|No|Number|`1000`|`1000`|The maximum number of records retrieved by the outbox worker from the outbox message table in each iteration. Max value: 10000.|

### Database configurations

|Field|Sub field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|-|
|`host`||Yes|String||`localhost`|Database host|
|`port`||Yes|Number|`3306`|`3306`|Database port|
|`username`||Yes|String||`outbox_worker`|Database username|
|`password`||Yes|String||`password`|Database password|
|`name`||Yes|String|`mysql`|`ap_production`|Database name|
|`sshTunnel`||No|Boolean|`false`|`false`|Whether or not the outbox worker connects to the database through an SSH tunnel|
|`tls`||No|Object|||TLS configurations|
||`insecureSkipVerify`|No|Boolean|`false`|`false`|Same as `InsecureSkipVerify` in [tls.Config](https://pkg.go.dev/crypto/tls#Config).|
||`serverName`|No|String|||Same as `ServerName` in [tls.Config](https://pkg.go.dev/crypto/tls#Config).|
||`caFile`|No|String|||CA file path|
|`maxOpenConn`||No|Number|`10`|`10`|The maximum number of open connections to the database|
|`maxLifeTimeSecond`||No|Number|`300`|`300`|The maximum amount of time a connection may be reused|
|`maxIdleConn`||No|Number|`1`|`1`|The maximum number of connections in the idle connection pool|
|`maxIdleSecond`||No|Number|`0`|`0`|The maximum amount of time a connection may be idle|

### SSH configurations

SSH configurations are required if `database.sshTunnel` is set to `true`.

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`privateKey`|Yes|String|||Private key file|
|`host`|Yes|String|||SSH server host|
|`port`|Yes|String|`22`|`22`|SSH server port|
|`username`|No|String|||SSH username|
|`hostKeyAlgorithms`|No|String[]|`["ssh-ed25519"]`||The public key algorithms that the client will accept from the server for host key authentication|
|`knownHosts`|No|String|`~/.ssh/known_hosts`||SSH host key file|

### Publisher configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`aws`|Yes|Object|||AWS configurations|
|`refetchTimer`|No|Object|||Refetch timer configuration|

AWS configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`accessKey`|No|String|||AWS access key ID|
|`secretKey`|No|String|||AWS secret access key|
|`region`|No|String|`ap-northeast-1`||AWS configurations|
|`sns`|Yes|Object|||SNS configurations|

SNS configurations

|Field|Sub field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|-|
|`endpoint`||No|String|||AWS endpoint|
|`resources`||No|Object|||Resource configurations|
||`whitelist`|No|String[]|||A list of allowed SNS topics. The outbox worker will send messages to only the SNS topics in this list.|
||`blacklist`|No|String[]|||A list of blocked SNS topics. The outbox worker will not send messages to the SNS topics in this list.|
||`skipEvents`|No|String[]|||Specify events to be skipped without sending to SNS|
||`ignoreFetchErrorResources`|No|String[]|||Specify resources to ignore in the event of an error when retrieving resource attributes.|

**Use skipEvents with caution.**

**When using `skipEvents`, please understand that event propagation is not sequential.**

Refetch timer configuration

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`enabled`|No|Boolean|`true`|`true`|Refetch timer is enabled|
|`interval`|No|String|`24h`|`24h`|Interval(duration format)|

### Logging configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`level`|No|String|`info`|`info`|Logging level. Available values: `debug`, `info`, `warn`, `error`.|
|`handler`|No|String or String[]|`text`|`text`|Logging handler. Available values: `text`, `json`, `rollbar`, `sentry`, `datadog`.|
|`rollbar`|Yes when `handler` is set to `rollbar`|Object|||Rollbar configurations|
|`sentry`|Yes when `handler` is set to `sentry`|Object|||Sentry configurations|

#### Rollbar configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`level`|No|String|`warn`|`warn`|Set log level to notify Rollbar. values: `debug`, `info`, `warn`, `error`|
|`token`|Yes|String||`xxx`|Set the server token for Rollbar|


#### Sentry configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`level`|No|String|`warn`|`warn`|Set log level to notify Sentry. values: `debug`, `info`, `warn`, `error`|
| `DSN`|Yes|String||`https://xxx@xxx.ingest.sentry.io`|Configure Sentry dsn|
|`sampleRate`|No|Float|1.0|1.0|Sampling Error Events|
|`sendDefaultPII`|No|Boolean|false||If this flag is enabled, certain personally identifiable information (PII) is added by active integrations.|

### Tracking configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`enabled`|No|Boolean|`false`|`false`|Whether or not tracking is enabled|
|`agentAddr`|Yes when `enabled` is set to `true`|String||`otel-collector:4317`|Tracking agent address|
|`serviceName`|No|String|`outbox-worker`|`outbox-worker`|Service name|
|`environment`|No|String|`prod`|`prod`|Environment|
