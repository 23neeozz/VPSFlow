module github.com/bosscloud/bosscloud/libs/go/security

go 1.23.0

require (
	github.com/bosscloud/bosscloud/libs/go/errors v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	golang.org/x/crypto v0.32.0
)

replace github.com/bosscloud/bosscloud/libs/go/errors => ../errors
