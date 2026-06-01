module github.com/jtenhave/not-just-noise/integrations

go 1.26.3

require (
	github.com/aws/aws-sdk-go-v2 v1.41.9
	github.com/aws/aws-sdk-go-v2/service/sns v1.39.17
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.29
	github.com/go-sql-driver/mysql v1.10.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/jtenhave/not-just-noise/contracts v0.0.0-00010101000000-000000000000
	github.com/jtenhave/not-just-noise/lib v0.0.0-20260527015339-467ad370d6ce
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.25 // indirect
	github.com/aws/smithy-go v1.26.0 // indirect
)

replace github.com/jtenhave/not-just-noise/contracts => ../contracts/
