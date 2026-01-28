package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	targetURL := os.Getenv("BASE_URL")
	if targetURL == "" {
		targetURL = "http://localhost:8080/api/v1/health/ready"
	} else {
		targetURL = targetURL + "/health/ready"
	}

	// Allow overriding duration via flag, default to 6 hours for endurance
	durationPtr := flag.Duration("duration", 6*time.Hour, "Duration of the endurance test")
	rpsPtr := flag.Int("rps", 200, "RPS to sustain")
	flag.Parse()

	fmt.Printf("Starting Endurance Test against %s\n", targetURL)
	fmt.Printf("Duration: %s, RPS: %d\n", *durationPtr, *rpsPtr)

	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    targetURL,
	})

	attacker := vegeta.NewAttacker()
	rate := vegeta.Rate{Freq: *rpsPtr, Per: time.Second}

	var metrics vegeta.Metrics
	// Print progress every 1 minute
	ticker := time.NewTicker(1 * time.Minute)
	startTime := time.Now()

	go func() {
		for range ticker.C {
			elapsed := time.Since(startTime)
			fmt.Printf("Elapsed: %s / %s | Current P99: %s | Success: %.2f%%\n",
				elapsed.Round(time.Second), *durationPtr, metrics.Latencies.P99, metrics.Success*100)
		}
	}()

	for res := range attacker.Attack(targeter, rate, *durationPtr, "Endurance Test") {
		metrics.Add(res)
	}
	ticker.Stop()
	metrics.Close()

	fmt.Println("\nEndurance Test Completed.")
	fmt.Printf("Final Results:\n Mean Latency: %s\n Max Latency: %s\n Success Rate: %.2f%%\n",
		metrics.Latencies.Mean, metrics.Latencies.Max, metrics.Success*100)
}
