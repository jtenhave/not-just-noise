module github.com/jtenhave/not-just-noise/audio-service

go 1.26.3

require (
	github.com/jtenhave/not-just-noise/contracts v1.0.0
	github.com/jtenhave/not-just-noise/dispatch-service v1.0.0
	github.com/jtenhave/not-just-noise/integrations v1.0.0
	github.com/jtenhave/not-just-noise/lib v1.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0 // indirect
)

replace github.com/jtenhave/not-just-noise/dispatch-service => ../dispatch-service/

replace github.com/jtenhave/not-just-noise/contracts => ../contracts/

replace github.com/jtenhave/not-just-noise/integrations => ../integrations/

replace github.com/jtenhave/not-just-noise/lib => ../lib/
