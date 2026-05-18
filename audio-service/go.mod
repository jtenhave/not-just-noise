module github.com/jtenhave/not-just-noise/audio-service

go 1.26.3

require github.com/jtenhave/not-just-noise/lib v1.0.0

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
)

require (
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0 // indirect
)

replace github.com/jtenhave/not-just-noise/lib => ../lib/
