package main

import (
	"flag"

	"github.com/jtenhave/not-just-noise/audio-service/internal/adapters"
	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/audio-service/internal/config"
	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/http"
	"github.com/jtenhave/not-just-noise/lib/transactionaljob"
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

	// Create audio repository
	audioRepo := audio.NewAudioRepo(mysql)

	// Create transactional job client
	transactionalJobClient := transactionaljob.NewTransactionalJobClient(mysql)

	// Create audio publish adapter
	audioPublishAdapter := adapters.NewPublishAdapter(config.SNS, transactionalJobClient)

	// Create audio service
	audioService := audio.NewAudioService(mysql, audioRepo, audioPublishAdapter)

	// Create audio routes
	audioRoutes := audio.CreateRoutes(audioService)

	// Start HTTP server
	http.StartServer(audioRoutes, *portPtr)
}
