package main

import (
	"context"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsSNS "github.com/aws/aws-sdk-go-v2/service/sns"
	awsSQS "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jtenhave/not-just-noise/dispatch-service/internal/config"
	"github.com/jtenhave/not-just-noise/dispatch-service/internal/dispatch"
	transactionaljob "github.com/jtenhave/not-just-noise/dispatch-service/internal/dispatch"
	"github.com/jtenhave/not-just-noise/dispatch-service/internal/dispatcher"
	"github.com/jtenhave/not-just-noise/integrations/mysql"
	"github.com/jtenhave/not-just-noise/integrations/sns"
	"github.com/jtenhave/not-just-noise/integrations/sqs"
)

func main() {
	// Load configuration
	config, err := config.LoadConfig("internal/config/config.json")
	if err != nil {
		panic(err)
	}

	// Create MySQL connection
	mysql, err := mysql.NewMySQLConnection(mysql.MySQLConfig{
		Host:     config.MySQL.Host,
		Port:     config.MySQL.Port,
		User:     config.MySQL.User,
		Password: config.MySQL.Password,
		DBName:   config.MySQL.DBName,
	})
	if err != nil {
		panic(err)
	}

	// Create dispatch repository
	dispatchRepo := dispatch.NewDispatchRepo(mysql)

	// Load AWS configuration
	awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	// Create AWS SNS client
	awsSNSClient := awsSNS.NewFromConfig(awsConfig)

	// Create SNS client
	snsClient := sns.NewSNSClient(awsSNSClient)

	// Create notify dispatch handler
	notifyDispatchHandler := dispatcher.NewNotifyDispatcher(snsClient)

	// Create AWS SQS client
	awsSQSClient := awsSQS.NewFromConfig(awsConfig)

	// Create SQS client
	sqsClient := sqs.NewSQSClient(awsSQSClient)

	// Create queue dispatch handler
	queueDispatchHandler := dispatcher.NewQueueDispatcher(sqsClient)

	// Create log dispatch handler
	logDispatchHandler := dispatcher.NewLogDispatcher()

	// Create dispatcher
	dsptchr := dispatcher.NewDispatcher()
	dsptchr.RegisterDispatcher(dispatcher.DispatcherTypeNotify, notifyDispatchHandler)
	dsptchr.RegisterDispatcher(dispatcher.DispatcherTypeQueue, queueDispatchHandler)
	dsptchr.RegisterDispatcher(dispatcher.DispatcherTypeLog, logDispatchHandler)

	// Create transactional job service
	dispatchService := dispatch.NewDispatchService(mysql, dispatchRepo, dsptchr)

	transactionalJobWorker := transactionaljob.NewDispatchWorker(transactionaljob.WorkerConfig{
		MaxWorkers:    config.Worker.MaxWorkers,
		MaxBatchSize:  config.Worker.MaxBatchSize,
		NoJobDelay:    config.Worker.NoJobDelay,
		JobBufferSize: config.Worker.JobBufferSize,
	}, dispatchService)

	// Start transactional job worker
	transactionalJobWorker.Start(context.Background())
}
