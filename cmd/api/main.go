package main

import (
	"log"
	"os"

	"github.com/braden0236/o11y-golang/internal/config"
	"github.com/braden0236/o11y-golang/internal/metric"
	"github.com/braden0236/o11y-golang/internal/server"
)

func main() {

	conf, err := config.InitConfig()
	if err != nil {
		log.Printf("Failed to initialize configuration: %v", err)
		os.Exit(100)
	}

	metrics := metric.NewMetrics(metric.WithBasicAuth(conf.Metrics.Username, conf.Metrics.Password))

	server := server.NewServer(metrics)

	go func() {
		if err := server.Run(conf.Server.Port); err != nil {
			log.Printf("Server start failed: %v", err)
			os.Exit(101)
		}
	}()

	if err := server.WaitForShutdown(nil); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
