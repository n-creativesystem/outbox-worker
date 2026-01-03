[English](README.md) | [日本語](README.ja.md)

# 概要

このライブラリは、Outboxメッセージテーブルから定期的にメッセージを取得し、設定されたバックエンドに配信する outbox worker を実装します。

対応バックエンド:

- AWS SNS
- NATS JetStream

# 使い方

## 仕組み

### 設定の読み込み

- Worker は YAML ファイルから設定を読み込みます。
- デフォルト値は自動的に適用されます。
- 環境変数展開に対応しています（Compose 風のテンプレート置換）。必要に応じて値をクォートしてください。
- バリデーションエラーは YAML のフィールド名で出力されます。

### ポーリングモデル

- ポーリング開始前に、Worker はバックエンド側のリソース（例: SNS topics）を取得します。
- `publisher.refetchTimer.enabled` が true の場合、バックエンド側リソースを定期的に再取得します。
- ポーリングループは `outbox.pollingInterval` 間隔で実行されます。
- 各イテレーションでは、`id` 昇順で最大 `outbox.findEventLimit` 件を読み込みます。
- レコードは `aggregate_type` 単位で並列処理されます。

### 順序制御とレート制限

- `aggregate_type` ごとに、レコードは `aggregate_id` でグルーピングされます。
- `aggregate_id` が空でない場合、同一 `aggregate_id` グループ内のレコードは `id` でソートされ、順序通りに publish されます。
- `aggregate_id` が空でないグループで publish に失敗した場合、順序を崩さないために、そのグループ内の残りのレコードの publish を停止します。
- レート制限は `aggregate_type` 単位で `outbox.throughput` により適用されます（最大 3000）。

### リトライ挙動

- レコードが **fetch 対象** となる条件:
  - `retry_count` が NULL、または
  - `retry_count < outbox.retryCount`
- `retry_at` が設定されていて未来時刻の場合、Worker はそのレコードを publish しません（"waiting for the retry" をログ出力）。
- publish に失敗した場合、Worker は指数バックオフで `retry_count` と `retry_at` を更新します:

  `nextRetryTime = now + retryBackoff * (2 ** retry_count)`

### `sent_at` と後処理

- publish 成功時、Worker は `sent_at` を更新します。
- 注意: 現実装では fetch 時に `sent_at` が設定済みの行を除外していません。再送を避けたい場合、アプリ側で publish 済み行（例: `sent_at IS NOT NULL`）を削除/アーカイブしてください。

### スキップ挙動

- `publisher.aws.sns.resources.skipEvents` は、fetch 時に `event` で除外します。
- `publisher.aws.sns.resources.whitelist` / `blacklist` は、fetch 時に `aggregate_type` でフィルタします。
- バックエンドの送信先を解決できない場合（例: SNS topic が見つからない等）、Worker は warn を出し、その行の retry フィールドを更新しません（後続のポーリングで再度試行されます）。

## Outbox メッセージテーブル

アプリケーションは、以下のカラムを持つ outbox メッセージテーブルを作成する必要があります。

|Column|Type|Description|Updater|
|-|-|-|-|
|`id`|bigint, PK|ID|app|
|`aggregate_type`|varchar(255), not null|配信先識別子（**ARN もしくは RN 形式が必須**）|app|
|`aggregate_id`|varchar(255), not null|メッセージグループID|app|
|`event`|varchar(255), not null|イベント種別。メッセージ属性/ヘッダー `Event` に設定されます。|app|
|`payload`|JSON, not null|メッセージ本文|app|
|`retry_at`|datetime, null|publish 失敗時の次回リトライ時刻|outbox worker|
|`retry_count`|int, null|リトライ回数|outbox worker|
|`sent_at`|datetime, null|publish 成功時刻|outbox worker|

### `aggregate_type` の形式（必須）

Outbox worker は `aggregate_type` の *service* 部分を元に、どのバックエンド publisher を使うかを選択します。
そのため `aggregate_type` は必ず以下のいずれかである必要があります:

- AWS ARN（例: SNS topic ARN）、または
- RN（ARN と同じセクション構造で、prefix が `rn:` のもの）

例:

- SNS: `arn:aws:sns:ap-northeast-1:123456789012:my-topic`
- SNS FIFO: `arn:aws:sns:ap-northeast-1:123456789012:my-topic.fifo`
- NATS: `rn::nats:::orders`

注意:

- backend のキーは、パースされた service 名（例: `sns`, `nats`）と一致している必要があります。
- `aggregate_type` が ARN/RN 形式でない場合、Worker が publisher を選択できず publish に失敗します。

### バックエンド別の挙動（詳細）

#### AWS SNS

- 宛先: ARN の resource 部分は SNS の topic 名として扱われます。
- Topic 解決:
  - 起動時（および refetch timer が有効なら定期的に）、Worker は SNS topics を列挙し、**topic 名**（ARN の `resource` 部分）でキャッシュします。
  - publish 時は `aggregate_type` に topic のフル ARN を指定できます。Worker はこれを topic 名（ARN の `resource`）に解決してキャッシュから Target を見つけます。
    - 例: `aggregate_type=arn:aws:sns:ap-northeast-1:123456789012:my-topic` は `my-topic` に解決されます。
- グルーピング: `aggregate_id` は `MessageGroupId` として使用されます。
- 属性: `event` / `producerName` は message attributes `Event` / `Producer` として送信されます。
- `ignoreFetchErrorResources`:
  - リソース探索中に特定 topic の属性取得（`GetTopicAttributes`）に失敗した場合、通常は warn を出してその topic をキャッシュしません。
  - topic 名が `publisher.aws.sns.resources.ignoreFetchErrorResources` に含まれている場合、warn を抑制します（ただし topic がキャッシュされない点は同じです）。

#### NATS JetStream

- 宛先:
  - `aggregate_type` が RN/ARN の場合、パースした **resource** 部分が topic 名として使用されます。
  - それ以外（必須形式に従っていれば通常発生しません）では、生の `aggregate_type` が使用されます。
- Subject 形式: `${topicName}.${aggregate_id}`
  - 例: `rn::nats:::orders` + `aggregate_id=1:1` => `orders.1:1`
- ヘッダー: `event` / `producerName` は headers `Event` / `Producer` として送信されます。
- JetStream MsgID: payload バイト列の SHA-256（16進文字列）で算出されます。

## ヘルスチェック

`polling` コマンドは HTTP のヘルスチェックサーバーを起動します。
バインドアドレスは `--health-check-addr` で変更できます（デフォルト: `:8080`）。

カラム名は現時点では固定ですが、将来的にアプリ側で動的にカラム名を定義できるようにする可能性があります。

アプリケーション側で outbox メッセージテーブルのマイグレーションを行ってください。

アプリケーション要件に応じて追加カラムを持たせても構いません。例:
  - `tenant_uid` や `office_id`: シャーディングや分析用途
  - `created_at`: 作成日時
  - `updated_at`: 更新日時
    - 注: outbox worker は `updated_at` を更新しないため、DB側で自動更新することを推奨します。

## 設定

YAML の設定ファイルで以下を定義できます。サンプルは [examples/config/outbox_polling.yaml](examples/config/outbox_polling.yaml) を参照してください。

|Field|Required|Description|
|-|-|-|
|`outbox`|Yes|Outbox worker 設定|
|`database`|Yes|DB 設定|
|`publisher`|Yes|Publisher 設定|
|`logging`|No|ログ設定|
|`tracking`|No|トレーシング設定|

### Outbox worker 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`schema`|No|String|||DB schema|
|`tableName`|Yes|String|`outbox`|`outbox_messages`|Outbox テーブル名|
|`producerName`|Yes|String||`NeoAP`|Producer 名。メッセージ属性/ヘッダー `Producer` に設定されます。|
|`pollingInterval`|No|Time duration|`5s`|`5s`|ポーリング間隔|
|`retryCount`|No|Number|`10`|`10`|publish 失敗時の最大リトライ回数|
|`retryBackoff`|No|Time duration|`20s`|`20s`|指数バックオフの基準時間（`nextRetryTime = now + retryBackoff * (2 ** retry_count)`）|
|`throughput`|No|Number|`3000`|`3000`|`aggregate_type` あたりの最大送信件数/秒（最大 3000）|
|`findEventLimit`|No|Number|`1000`|`1000`|1回の fetch 件数上限（最大 10000）|

### Database 設定

|Field|Sub field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|-|
|`uri`||Yes|String||`mysql://admin:pass@tcp(localhost:3306)/app?parseTime=true`|DB 接続文字列。`mysql://` / `postgres://` に対応。|
|`username`||Yes|String||`outbox_worker`|DB ユーザー名（サポートされる場合 `uri` の user 部分を上書き）|
|`password`||Yes|String||`password`|DB パスワード（サポートされる場合 `uri` の password 部分を上書き）|
|`tls`||No|Object|||TLS 設定（MySQL のみ）|
||`insecureSkipVerify`|No|Boolean|`false`|`false`|[tls.Config](https://pkg.go.dev/crypto/tls#Config) の `InsecureSkipVerify` 相当|
||`serverName`|No|String|||[tls.Config](https://pkg.go.dev/crypto/tls#Config) の `ServerName` 相当|
||`caFile`|No|String|||CA ファイルパス|
|`ssh`||No|Object|||SSH トンネル設定（MySQL のみ）|
|`maxOpenConn`||No|Number|`10`|`10`|最大オープンコネクション数|
|`maxLifeTimeSecond`||No|Number|`300`|`300`|コネクションの最大再利用時間（秒）|
|`maxIdleConn`||No|Number|`1`|`1`|最大アイドルコネクション数|
|`maxIdleSecond`||No|Number|`0`|`0`|最大アイドル時間（秒）|

### SSH 設定

`database.ssh` を設定した場合に使用されます。

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`privateKey`|Yes|String|||秘密鍵ファイル|
|`host`|Yes|String|||SSH サーバーホスト|
|`port`|No|Number|`22`|`22`|SSH サーバーポート|
|`username`|No|String|||SSH ユーザー名（多くのSSHサーバーでは必須）|
|`hostKeyAlgorithms`|No|String[]|`["ssh-ed25519"]`||ホスト鍵アルゴリズム|
|`knownHosts`|No|String|||known_hosts ファイル（OpenSSH 形式）|

### Publisher 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`aws`|No|Object|||AWS 設定（SNS publisher）|
|`nats`|No|Object|||NATS 設定（JetStream publisher）|
|`refetchTimer`|No|Object|||リソース再取得タイマー|

`publisher.aws` または `publisher.nats` のいずれか（または両方）を設定してください。

AWS 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`accessKey`|No|String|||AWS access key ID|
|`secretKey`|No|String|||AWS secret access key|
|`region`|No|String|`ap-northeast-1`||AWS region|
|`sns`|Yes|Object|||SNS 設定|

SNS 設定

|Field|Sub field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|-|
|`endpoint`||No|String|||AWS endpoint|
|`resources`||No|Object|||リソース設定|
||`whitelist`|No|String[]|||許可する SNS topic 名のリスト|
||`blacklist`|No|String[]|||拒否する SNS topic 名のリスト|
||`skipEvents`|No|String[]|||送信せずにスキップする event のリスト|
||`ignoreFetchErrorResources`|No|String[]|||リソース属性取得エラーを無視（warn抑制）する topic 名のリスト|

**skipEvents は注意して使用してください。**

**skipEvents を使う場合、イベント伝播は sequential ではありません。**

Refetch timer 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`enabled`|No|Boolean|`true`|`true`|有効/無効|
|`interval`|No|String|`24h`|`24h`|再取得間隔（duration形式）|

NATS 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`server`|Yes|String||`nats://localhost:4222`|NATS server URL|
|`clientName`|No|String|||クライアント名|
|`maxReconnects`|No|Number|`5`|`5`|最大再接続回数|
|`reconnectWait`|No|Time duration|`2s`|`2s`|再接続待ち|
|`reconnectJitter`|No|Time duration|`100ms`|`100ms`|jitter|
|`reconnectJitterTLS`|No|Time duration|`1s`|`1s`|TLS用jitter|
|`timeout`|No|Time duration|`5s`|`5s`|接続タイムアウト|
|`domain`|No|String|||JetStream domain|
|`tls`|No|Object|||TLS 設定|
|`userInfo`|No|Object|||ユーザー/パスワード認証|
|`userCredentials`|No|Object|||credentials 認証|
|`token`|No|String|||token 認証|

### Logging 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`level`|No|String|`info`|`info`|ログレベル（`debug`/`info`/`warn`/`error`）|
|`handler`|No|String or String[]|`text`|`text`|ハンドラ（`text`/`json`/`rollbar`/`sentry`/`datadog`）|
|`rollbar`|Yes when `handler` is set to `rollbar`|Object|||Rollbar 設定|
|`sentry`|Yes when `handler` is set to `sentry`|Object|||Sentry 設定|

#### Rollbar 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`level`|No|String|`warn`|`warn`|Rollbar に通知するログレベル|
|`token`|Yes|String||`xxx`|Rollbar server token|

#### Sentry 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`level`|No|String|`warn`|`warn`|Sentry に通知するログレベル|
| `DSN`|Yes|String||`https://xxx@xxx.ingest.sentry.io`|Sentry DSN|
|`sampleRate`|No|Float|1.0|1.0|Sampling rate|
|`sendDefaultPII`|No|Boolean|false||PII を送るか|

### Tracking 設定

|Field|Required|Data type|Default|Example|Description|
|-|-|-|-|-|-|
|`enabled`|No|Boolean|`false`|`false`|有効/無効|
|`agentAddr`|Yes when `enabled` is set to `true`|String||`otel-collector:4317`|トレーシングエージェントアドレス|
|`serviceName`|No|String|`outbox-worker`|`outbox-worker`|サービス名|
|`environment`|No|String|`prod`|`prod`|環境名|

## CLI

### Polling

ポーリング worker を起動します:

```bash
outbox-messenger outbox polling -c config.yaml --health-check-addr :8080
```
