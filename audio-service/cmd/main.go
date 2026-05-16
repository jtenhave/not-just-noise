package main

import (
	"flag"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/http"
)

func main() {
	portPtr := flag.Int("port", 8080, "the port to listen on")
	flag.Parse()

	audioService := audio.NewAudioService()
	audioRoutes := audio.CreateRoutes(audioService)

	http.StartServer(audioRoutes, *portPtr)
}
