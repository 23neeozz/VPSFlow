module github.com/vpsflow/vpsflow/services/auth

go 1.23.0

require (
	github.com/vpsflow/vpsflow/libs/go/config v0.0.0
	github.com/vpsflow/vpsflow/libs/go/errors v0.0.0
	github.com/vpsflow/vpsflow/libs/go/httpx v0.0.0
	github.com/vpsflow/vpsflow/libs/go/ids v0.0.0
	github.com/vpsflow/vpsflow/libs/go/observability v0.0.0
	github.com/vpsflow/vpsflow/libs/go/security v0.0.0
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/pquerna/otp v1.4.0
	github.com/prometheus/client_golang v1.20.5
	go.opentelemetry.io/otel v1.34.0
)

replace (
	github.com/vpsflow/vpsflow/libs/go/config => ../../libs/go/config
	github.com/vpsflow/vpsflow/libs/go/errors => ../../libs/go/errors
	github.com/vpsflow/vpsflow/libs/go/httpx => ../../libs/go/httpx
	github.com/vpsflow/vpsflow/libs/go/ids => ../../libs/go/ids
	github.com/vpsflow/vpsflow/libs/go/observability => ../../libs/go/observability
	github.com/vpsflow/vpsflow/libs/go/security => ../../libs/go/security
)
