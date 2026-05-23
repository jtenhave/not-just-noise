package main

import (
	"flag"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audioing"
	"github.com/jtenhave/not-just-noise/audio-service/internal/config"
	"github.com/jtenhave/not-just-noise/audio-service/internal/infrastructure/notify"
	"github.com/jtenhave/not-just-noise/audio-service/internal/infrastructure/repo"
	"github.com/jtenhave/not-just-noise/lib/database/mysql"
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
	mysql, err := mysql.NewConnectionQueryRunner(config.MySQL)
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
	audioRepo := repo.NewConnectionRepository(mysql)

	// Create audio notifier
	//audioNotifier := audioNotify.NewAudioNotifier(snsPublisher)

	// Create audio notify formatter
	audioNotifyFormatter := notify.NewAudioNotifyFormatter()

	// Create audio service
	audioService := audioing.NewAudioService(audioRepo, audioNotifyFormatter)

	// Create audio routes
	audioRoutes := audioing.CreateRoutes(audioService)

	// Start HTTP server
	http.StartServer(audioRoutes, *portPtr)
}
