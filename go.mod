module github.com/n-creativesystem/outbox-worker

go 1.25

require (
	entgo.io/ent v0.12.5
	github.com/aws/aws-sdk-go-v2 v1.23.2
	github.com/aws/aws-sdk-go-v2/config v1.25.8
	github.com/aws/aws-sdk-go-v2/credentials v1.16.6
	github.com/aws/aws-sdk-go-v2/service/sns v1.25.5
	github.com/aws/aws-sdk-go-v2/service/sqs v1.28.4
	github.com/cockroachdb/errors v1.12.0
	github.com/compose-spec/compose-go/v2 v2.10.0
	github.com/creasty/defaults v1.7.0
	github.com/doug-martin/goqu/v9 v9.19.0
	github.com/getsentry/sentry-go v0.40.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/goccy/go-yaml v1.11.2
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/jmoiron/sqlx v1.3.5
	github.com/n-creativesystem/go-packages/lib/trace v1.0.0
	github.com/nats-io/nats-server/v2 v2.12.3
	github.com/nats-io/nats.go v1.48.0
	github.com/rollbar/rollbar-go v1.4.8
	github.com/samber/slog-rollbar/v2 v2.1.0
	github.com/samber/slog-sentry/v2 v2.10.1
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	github.com/testcontainers/testcontainers-go v0.40.0
	github.com/testcontainers/testcontainers-go/modules/mysql v0.40.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.40.0
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.2.3
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.46.1
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
	go.uber.org/mock v0.6.0
	go.uber.org/multierr v1.11.0
	golang.org/x/crypto v0.46.0
	golang.org/x/sync v0.19.0
	golang.org/x/time v0.14.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/antithesishq/antithesis-sdk-go v0.5.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.25.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.8.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.17.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.20.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.25.6 // indirect
	github.com/aws/smithy-go v1.17.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/redact v1.1.6 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v28.5.2+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-playground/validator/v10 v10.15.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.4-0.20251030100505-070ab1a87a76 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.2.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/nats-io/jwt/v2 v2.8.0 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	github.com/shirou/gopsutil/v4 v4.25.11 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.64.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251213004720-97cd9d5aeac2 // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool go.uber.org/mock/mockgen
