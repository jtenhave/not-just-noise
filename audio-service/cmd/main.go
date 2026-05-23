package main

import (
	"flag"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/audio-service/internal/config"
	"github.com/jtenhave/not-just-noise/audio-service/internal/transactionaloutbox"
	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/http"
)

func main() {
	portPtr := flag.Int("port", 8080, "the port to listen on")
	flag.Parse()

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

	// Load AWS configuration
	/*awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}*/

	// Create SNS publisher
	//snsPublisher := notify.NewSNSPublisher(awsConfig, config.SNS.TopicArn)

	// Create audio repository
	audioRepo := audio.NewAudioRepo(mysql)

	// Create transactional outbox repository
	transactionalOutboxRepo := transactionaloutbox.NewTransactionalOutboxRepo(mysql)

	// Create audio service
	audioService := audio.NewAudioService(mysql, audioRepo, transactionalOutboxRepo)

	// Create audio routes
	audioRoutes := audio.CreateRoutes(audioService)

	// Start HTTP server
	http.StartServer(audioRoutes, *portPtr)
}
