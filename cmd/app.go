package main

import (
	"github.com/joho/godotenv"
	"loadBalancer/internal"
	"log"
	"os"
	"strings"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Ошибка загрузки файла .env:", err)
	}

	loadBalancerStrategy := os.Getenv("LOAD_BALANCE_STRATEGY")
	servers := strings.Split(os.Getenv("SERVERS"), ",")
	healthCheckInterval := os.Getenv("HEALTH_CHECK_INTERVAL")
	serverPort := os.Getenv("SERVER_PORT")

	internal.StartServer(loadBalancerStrategy, servers, healthCheckInterval, serverPort)
}
