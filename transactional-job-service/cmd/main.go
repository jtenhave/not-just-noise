package main

import (
	"context"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/notify"
	"github.com/jtenhave/not-just-noise/transactional-job-service/internal/config"
	"github.com/jtenhave/not-just-noise/transactional-job-service/internal/transactionaljob"
	"github.com/jtenhave/not-just-noise/transactional-job-service/internal/transactionaljobhandlers"
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

	// Load AWS configuration
	awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	// Create SNS publisher
	snsPublisher := notify.NewSNSPublisher(awsConfig)

	// Create handlers
	handlers := map[string]transactionaljob.TransactionalJobHandler{
		"log":    transactionaljobhandlers.NewLogHandler(),
		"notify": transactionaljobhandlers.NewNotifyHandler(snsPublisher),
	}

	// Create transactional job service
	transactionalJobService := transactionaljob.NewTransactionalJobService(mysql, transactionalJobRepo, handlers)

	// Create audio routes
	transactionalJobWorker := transactionaljob.NewTransactionalJobWorker(config.Worker, transactionalJobService)

	// Start transactional job worker
	transactionalJobWorker.Start(context.Background())
}
