[English](README.md) | [日本語](README.ja.md)

# Overview

This library implements an outbox worker which periodically fetches messages from an outbox message table and publishes them to a configured backend.

Currently supported backends:

- AWS SNS
- NATS JetStream

# Usage

## How it works

### Config loading

- The worker loads configuration from a YAML file.
- Default values are applied automatically.
- Environment variable expansion is supported (Compose-style template substitution), remember to quote values when needed.
- Validation errors use YAML field names.

### Polling model

- Before polling starts, the worker fetches backend resources (e.g. SNS topics).
- If `publisher.refetchTimer.enabled` is true, the worker periodically refetches backend resources.
- The polling loop runs every `outbox.pollingInterval`.
- Each iteration reads up to `outbox.findEventLimit` records ordered by `id` ascending.
- Records are processed in parallel per `aggregate_type`.

### Ordering and rate limiting

- For each `aggregate_type`, records are grouped by `aggregate_id`.
- If `aggregate_id` is not empty, records in the same `aggregate_id` group are sorted by `id` and published in order.
- If publishing fails for a non-empty `aggregate_id` group, the worker stops publishing the remaining records in that group (to avoid breaking ordering).
- Rate limiting is applied per `aggregate_type` using `outbox.throughput` (max 3000).

### Retry behavior

- A record is **eligible for fetch** when:
  - `retry_count` is NULL, or
  - `retry_count < outbox.retryCount`
- If `retry_at` is set and is in the future, the worker does **not** publish the record yet (it logs "waiting for the retry").
- When publishing fails, the worker updates `retry_count` and `retry_at` using exponential backoff:

  `nextRetryTime = now + retryBackoff * (2 ** retry_count)`

### `sent_at` and cleanup

- On successful publish, the worker updates `sent_at`.
- Note: the current implementation does **not** filter out rows with `sent_at` set when fetching. If you do not want re-sends, your application should delete/archive published rows (e.g. where `sent_at IS NOT NULL`).

### Skipping behavior

- `publisher.aws.sns.resources.skipEvents` excludes rows by `event` at fetch time.
- `publisher.aws.sns.resources.whitelist` / `blacklist` filters rows by `aggregate_type` at fetch time.
- If the backend destination cannot be resolved (e.g. SNS topic not found), the worker logs a warning and does not update retry fields for that row (it will be attempted again in later polling runs).

## Outbox message table

Applications MUST create an outbox message table with the following columns:

|Column|Type|Description|Updater|
|-|-|-|-|
|`id`|bigint, PK|ID|app|
|`aggregate_type`|varchar(255), not null|Publish destination identifier (**must be ARN or RN format**)|app|
|`aggregate_id`|varchar(255), not null|Message group ID|app|
|`event`|varchar(255), not null|Event type. This value will be set to a message attribute/header `Event`.|app|
|`payload`|JSON, not null|Message body|app|
|`retry_at`|datetime, null|Next retry time when the outbox worker fails to publish the message|outbox worker|
|`retry_count`|int, null|The number of retry attempts|outbox worker|
|`sent_at`|datetime, null|The date and time when the message was published successfully|outbox worker|

### `aggregate_type` format (required)

The outbox worker selects a backend publisher based on the *service* part of `aggregate_type`.
Therefore, `aggregate_type` MUST be either:

- an AWS ARN (e.g. SNS topic ARN), or
- an RN (the same section format as ARN, with the `rn:` prefix)

Examples:

- SNS: `arn:aws:sns:ap-northeast-1:123456789012:my-topic`
- SNS FIFO: `arn:aws:sns:ap-northeast-1:123456789012:my-topic.fifo`
- NATS: `rn::nats:::orders`

Notes:

- The backend key must match the parsed service name (e.g. `sns`, `nats`).
- If `aggregate_type` is not ARN/RN format, the worker cannot choose a publisher and will fail to publish.

### Backend behavior (details)

#### AWS SNS

- Destination: the ARN resource part is treated as the SNS topic name.
- Topic resolution:
  - On startup (and optionally by refetch timer), the worker lists SNS topics and caches them by **topic name** (the ARN `resource` part).
  - When publishing, `aggregate_type` can be the full topic ARN; the worker resolves it to the topic name (ARN `resource`) to find the cached target.
    - Example: `aggregate_type=arn:aws:sns:ap-northeast-1:123456789012:my-topic` resolves to `my-topic`.
- Grouping: `aggregate_id` is used as `MessageGroupId`.
- Attributes: `event` and `producerName` are sent as message attributes `Event` and `Producer`.
- `ignoreFetchErrorResources`:
  - When the worker fails to retrieve attributes for a topic during resource discovery (`GetTopicAttributes`), it normally logs a warning and skips caching that topic.
  - If the topic name is listed in `publisher.aws.sns.resources.ignoreFetchErrorResources`, the warning is suppressed (the topic is still skipped).

#### NATS JetStream

- Destination:
  - If `aggregate_type` is RN/ARN, the parsed **resource** part is used as the topic name.
  - Otherwise (should not happen if you follow the required format), the raw `aggregate_type` would be used.
- Subject format: `${topicName}.${aggregate_id}`
  - Example: `rn::nats:::orders` + `aggregate_id=1:1` => `orders.1:1`
- Headers: `event` and `producerName` are sent as headers `Event` and `Producer`.
- JetStream MsgID: computed as SHA-256 of the payload bytes (hex string).

## Health check

The polling command starts an HTTP health check server.
Use `--health-check-addr` to change the bind address (default: `:8080`).

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
|`publisher`|Yes|Publisher configurations|
|`logging`|No|Logging configurations|
|`tracking`|No|Tracking configurations|

### Outbox worker configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`schema`|No|String|||Database schema|
|`tableName`|Yes|String|`outbox`|`outbox_messages`|Outbox message table name|
|`producerName`|Yes|String||`example`|Producer name. This value will be set to a message attribute/header `Producer`.|
|`pollingInterval`|No|Time duration|`5s`|`5s`|The time interval between consecutive polling attempts|
|`retryCount`|No|Number|`10`|`10`|The maximum number of retry attempts that the outbox worker will make for a message when it fails to send the message to the backend.|
|`retryBackoff`|No|Time duration|`20s`|`20s`|The fundamental time duration used in calculating the next retry time by the formula `nextRetryTime = currentTime + retryBackoff * (2 ** currentRetryCount)`.|
|`throughput`|No|Number|`3000`|`3000`|The maximum number of messages to send per second per `aggregate_type`. Max value: 3000.|
|`findEventLimit`|No|Number|`1000`|`1000`|The maximum number of records retrieved by the outbox worker from the outbox message table in each iteration. Max value: 10000.|

### Database configurations

|Field|Sub field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|-|
|`uri`||Yes|String||`mysql://admin:pass@tcp(localhost:3306)/app?parseTime=true`|Database connection string. Supported schemes: `mysql://` and `postgres://`.|
|`username`||Yes|String||`outbox_worker`|Database username (overrides the user part in `uri` when supported)|
|`password`||Yes|String||`password`|Database password (overrides the password part in `uri` when supported)|
|`tls`||No|Object|||TLS configurations (MySQL only)|
||`insecureSkipVerify`|No|Boolean|`false`|`false`|Same as `InsecureSkipVerify` in [tls.Config](https://pkg.go.dev/crypto/tls#Config).|
||`serverName`|No|String|||Same as `ServerName` in [tls.Config](https://pkg.go.dev/crypto/tls#Config).|
||`caFile`|No|String|||CA file path|
|`ssh`||No|Object|||SSH tunnel configurations (MySQL only)|
|`maxOpenConn`||No|Number|`10`|`10`|The maximum number of open connections to the database|
|`maxLifeTimeSecond`||No|Number|`300`|`300`|The maximum amount of time a connection may be reused|
|`maxIdleConn`||No|Number|`1`|`1`|The maximum number of connections in the idle connection pool|
|`maxIdleSecond`||No|Number|`0`|`0`|The maximum amount of time a connection may be idle|

### SSH configurations

SSH configurations are used when `database.ssh` is configured.

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`privateKey`|Yes|String|||Private key file|
|`host`|Yes|String|||SSH server host|
|`port`|No|Number|`22`|`22`|SSH server port|
|`username`|No|String|||SSH username (required by most SSH servers)|
|`hostKeyAlgorithms`|No|String[]|`["ssh-ed25519"]`||The public key algorithms that the client will accept from the server for host key authentication|
|`knownHosts`|No|String|||SSH host key file (OpenSSH known_hosts format)|

### Publisher configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`aws`|No|Object|||AWS configurations (SNS publisher)|
|`nats`|No|Object|||NATS configurations (JetStream publisher)|
|`refetchTimer`|No|Object|||Refetch timer configuration|

At least one of `publisher.aws` or `publisher.nats` must be configured.

AWS configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`accessKey`|No|String|||AWS access key ID|
|`secretKey`|No|String|||AWS secret access key|
|`region`|No|String|`ap-northeast-1`||AWS region|
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

NATS configurations

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`server`|Yes|String||`nats://localhost:4222`|NATS server URL|
|`clientName`|No|String|||NATS client name|
|`maxReconnects`|No|Number|`5`|`5`|Maximum reconnect attempts|
|`reconnectWait`|No|Time duration|`2s`|`2s`|Reconnect wait|
|`reconnectJitter`|No|Time duration|`100ms`|`100ms`|Reconnect jitter|
|`reconnectJitterTLS`|No|Time duration|`1s`|`1s`|Reconnect jitter for TLS|
|`timeout`|No|Time duration|`5s`|`5s`|Connection timeout|
|`domain`|No|String|||JetStream domain|
|`tls`|No|Object|||TLS configurations|
|`userInfo`|No|Object|||User/password authentication|
|`userCredentials`|No|Object|||NATS credentials authentication|
|`token`|No|String|||Token authentication|

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

## CLI

### Polling

Run the polling worker:

```bash
outbox-messenger outbox polling -c config.yaml --health-check-addr :8080
```
