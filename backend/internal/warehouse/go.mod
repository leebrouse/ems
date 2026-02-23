module github.com/leebrouse/ems/backend/warehouse

go 1.24.11

replace github.com/leebrouse/ems/backend/internal/common => ../common

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/leebrouse/ems/backend/internal/common v0.0.0
	google.golang.org/grpc v1.79.1
)
