module github.com/bosscloud/bosscloud/services/cluster

go 1.23.0

require (
	github.com/bosscloud/bosscloud/libs/go/config v0.0.0
	github.com/bosscloud/bosscloud/libs/go/errors v0.0.0
	github.com/bosscloud/bosscloud/libs/go/httpx v0.0.0
	github.com/bosscloud/bosscloud/libs/go/ids v0.0.0
	github.com/bosscloud/bosscloud/libs/go/observability v0.0.0
	github.com/bosscloud/bosscloud/libs/go/security v0.0.0
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/prometheus/client_golang v1.20.5
	github.com/valyala/fasthttp v1.51.0
)

replace (
	github.com/bosscloud/bosscloud/libs/go/config => ../../libs/go/config
	github.com/bosscloud/bosscloud/libs/go/errors => ../../libs/go/errors
	github.com/bosscloud/bosscloud/libs/go/httpx => ../../libs/go/httpx
	github.com/bosscloud/bosscloud/libs/go/ids => ../../libs/go/ids
	github.com/bosscloud/bosscloud/libs/go/observability => ../../libs/go/observability
	github.com/bosscloud/bosscloud/libs/go/security => ../../libs/go/security
)
