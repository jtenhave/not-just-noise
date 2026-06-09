package main

import (
	"context"
	"net/http"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsSQS "github.com/aws/aws-sdk-go-v2/service/sqs"
	dispatchClient "github.com/jtenhave/not-just-noise/dispatch-service/client"
	"github.com/jtenhave/not-just-noise/integrations/mysql"
	"github.com/jtenhave/not-just-noise/integrations/s3"
	"github.com/jtenhave/not-just-noise/integrations/sqs"
	"github.com/jtenhave/not-just-noise/lib/log"
	adapters "github.com/jtenhave/not-just-noise/upload-service/adpaters"
	"github.com/jtenhave/not-just-noise/upload-service/internal/config"
	"github.com/jtenhave/not-just-noise/upload-service/internal/upload"
)

func main() {
	// Load configuration
	config, err := config.LoadConfig("internal/config/config.json")
	if err != nil {
		panic(err)
	}

	// Create logger
	ctx := log.SetupLogger(context.Background())

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

	// Create upload repository
	uploadRepo := upload.NewUploadRepo(mysql)

	// Create http client
	httpClient := http.DefaultClient

	// Create file downloader
	fileDownloader := adapters.NewFileDownloader(httpClient)

	// Load AWS configuration
	awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	// Create AWS S3 client
	awsS3Client := awsS3.NewFromConfig(awsConfig)

	// Create S3 client
	tempStorageClient := s3.NewS3Client(awsS3Client, config.S3.TempBucket)

	// Create dispatch client
	dispatchClient := dispatchClient.NewDispatchClient(mysql)

	// Create upload service
	uploadService := upload.NewUploadService(mysql, uploadRepo, fileDownloader, tempStorageClient, dispatchClient, config.SQS.UploadCommitQueueURL)

	// Create AWSSQS client
	awsSQSClient := awsSQS.NewFromConfig(awsConfig)

	// Create SQS client
	sqsClient := sqs.NewSQSClient(awsSQSClient)

	// Create upload handler
	uploadHandler := upload.NewUploadHandler(uploadService)

	// Create upload worker
	uploadWorker := upload.NewUploadWorker(upload.WorkerConfig{
		MaxWorkers:    config.Worker.MaxWorkers,
		MaxBatchSize:  config.Worker.MaxBatchSize,
		NoJobDelay:    config.Worker.NoJobDelay,
		JobBufferSize: config.Worker.JobBufferSize,
	}, sqsClient, config.SQS.UploadQueueURL, uploadHandler.HandleAudioChangedEventMessage)

	// Start upload worker
	uploadWorker.Start(ctx)
}
