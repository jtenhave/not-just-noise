package main

import (
	"flag"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/audio-service/internal/config"
	dispatchClient "github.com/jtenhave/not-just-noise/dispatch-service/client"
	"github.com/jtenhave/not-just-noise/integrations/mysql"
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

	// Create audio repository
	audioRepo := audio.NewAudioRepo(mysql)

	// Create dispatch client
	dispatchClient := dispatchClient.NewDispatchClient(mysql)

	// Create audio service
	audioService := audio.NewAudioService(mysql, audioRepo, dispatchClient, config.SNS.TopicArn)

	// Create audio routes
	audioRoutes := audio.CreateRoutes(audioService)

	// Start HTTP server
	http.StartServer(audioRoutes, *portPtr)
}
