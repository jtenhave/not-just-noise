package main

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/http"
	"github.com/jtenhave/not-just-noise/transactional-job-service/internal/config"
	"github.com/jtenhave/not-just-noise/transactional-job-service/internal/transactionaljob"
)

func main() {

	// Load configuration
	config, err := config.LoadConfig("internal/config/")
	if err != nil {
		panic(err)
	}

	// Create MySQL connection
	mysql, err := database.NewMySQLConnection(config.MySQL)
	if err != nil {
		panic(err)
	}

	// Create transactional job repository
	transactionalJobRepo := transactionaljob.NewTransactionalJobRepo(mysql)

	// Create HTTP client
	httpClient := http.NewClient()

	// Create transactional job service
	transactionalJobService := transactionaljob.NewTransactionalJobService(mysql, transactionalJobRepo, httpClient)

	// Create audio routes
	transactionalJobWorker := transactionaljob.NewTransactionalJobWorker(config.Worker, transactionalJobService)

	// Start transactional job worker
	transactionalJobWorker.Start(context.Background())
}
