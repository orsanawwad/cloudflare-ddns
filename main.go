package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/orsanawwad/cloudflare-ddns/cloudflare"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found.")
	}
	cfKey := os.Getenv("CFKEY")
	cfUser := os.Getenv("CFUSER")
	cfZone := os.Getenv("CFZONE")
	cfHosts := os.Getenv("CFHOSTS")
	hosts := strings.Split(cfHosts, " ")
	tickTimeValue := os.Getenv("TICKTIME")

	tickTime, err := time.ParseDuration(tickTimeValue)

	if err != nil {
		log.Fatal("Wrong Time Format")
	}

	// hosts := []string{"mc.crofis.net"}
	client := cloudflare.NewClient(
		cfKey,
		cfUser,
		cfZone,
		hosts,
		nil,
	)
	client.CheckAndUpdate()
	burstyLimiter := make(chan time.Time, 3)
	go func() {
		for t := range time.Tick(tickTime) {
			burstyLimiter <- t
		}
	}()

	for {
		<-burstyLimiter
		client.CheckAndUpdate()
		// fmt.Println("request", req, time.Now())
	}
}
