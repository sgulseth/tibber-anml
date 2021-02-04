package main

import (
	"os"
	"context"
	"log"

	"github.com/sgulseth/tibber-anml/pkg/draw"
	"github.com/sgulseth/tibber-anml/pkg/tibber"
)

var (
	tibberHomeID = os.Getenv("TIBBER_HOME_ID")
)

func main() {
	ctx := context.Background()

	ch := make(chan tibber.LiveMeasurement)
	go func() {
		err := tibber.Subscribe(ctx, tibberHomeID, ch)

		if err != nil {
			log.Fatalf("Unable to subscribe to id: %+v", err)
		}
	}()

	d := draw.Draw{}

	for measurement := range ch {
		// log.Printf("Time: %s Power: %d", measurement.Timestamp.Format(time.RFC3339), measurement.Power)

		d.Append(measurement.Timestamp, float64(measurement.Power))
		d.Flush()
	}
}
