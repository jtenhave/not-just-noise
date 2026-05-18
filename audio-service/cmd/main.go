package main

import (
	"flag"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/audio-service/internal/config"
	"github.com/jtenhave/not-just-noise/audio-service/internal/infrastructure/repo"
	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/http"
)

func main() {
	portPtr := flag.Int("port", 8080, "the port to listen on")
	flag.Parse()

	config, err := config.LoadConfig("internal/config/")
	if err != nil {
		panic(err)
	}

	mysql, err := database.NewMySQL(config.MySQL)
	if err != nil {
		panic(err)
	}

	audioRepo := repo.NewAudioRepo(mysql)
	audioService := audio.NewAudioService(audioRepo)
	audioRoutes := audio.CreateRoutes(audioService)

	http.StartServer(audioRoutes, *portPtr)
}
