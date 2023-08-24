package main

import (
	"math/rand"
	"time"

	_ "go.uber.org/automaxprocs"

	daemon "example.com/demo/internal/daemon"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// os.Setenv("ENV_REDIS_HOST", "127.0.0.1")
	// os.Setenv("ENV_REDIS_PORT", "6379")
	// os.Setenv("ENV_REDIS_PASSWORD", "")
	// os.Setenv("ENV_QUEUE_REGION", "HK")
	// os.Setenv("ENV_APP_MODE", "dev")

	daemon.NewApp("daemon").Run()
}
