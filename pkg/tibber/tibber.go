package tibber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zaba505/gws"
)

var (
	apiToken = os.Getenv("TIBBER_API_TOKEN")
	apiURL   = "wss://api.tibber.com/v1-beta/gql/subscriptions"

	client gws.Client
)

func init() {
	if apiToken == "" {
		log.Fatal("Missing api token")
	}

	ctx := context.Background()

	conn, err := gws.Dial(ctx, apiURL)
	if err != nil {
		log.Fatalf("gws.Dial: %+v", err)
	}

	client = gws.NewClient(conn, apiToken)
}

var liveMeasurementQuery = `subscription($homeID: ID!) {
	liveMeasurement(homeId:$homeID){
		timestamp
		power
		accumulatedConsumption
		accumulatedCost
		currency
		minPower
		averagePower
		maxPower
	}
}
`

type LiveMeasurement struct {
	Timestamp              time.Time `json:"timestamp"`
	Power                  int       `json:"power"`
	AccumulatedConsumption float32   `json:"accumulatedConsumption"`
	AccumulatedCost        float32   `json:"accumulatedCost"`
	Currency               string    `json:"currency"`
	MinPower               int32     `json:"minPower"`
	AveragePower           float32   `json:"averagePower"`
	MaxPower               int32     `json:"maxPower"`
}

type liveMeasurementResponse struct {
	LiveMeasurement LiveMeasurement `json:"liveMeasurement"`
}

// Subscribe to updates about kW/h and NOK for a given home ID, it will block the current routine until cancelled
func Subscribe(ctx context.Context, homeID string, ch chan LiveMeasurement) error {
	req := &gws.Request{
		Query: liveMeasurementQuery,
		Variables: map[string]interface{}{
			"homeID": homeID,
		},
	}

	subscription, err := client.Subscribe(ctx, req)

	if err != nil {
		return errors.Wrap(err, "client.Subscribe")
	}

	for {
		response, err := subscription.Recv(ctx)

		if err != nil {
			if err == gws.ErrUnsubscribed { // Tibber disconnects after 3 minutes~ so reconnect!
				return Subscribe(ctx, homeID, ch)
			}
			return errors.Wrap(err, "subscription.Recv")
		}

		if len(response.Errors) > 0 {
			var errors []string
			for _, err := range response.Errors {
				errors = append(errors, string(err))
			}

			return fmt.Errorf("response errors: %s", strings.Join(errors, "\n"))
		}

		var data liveMeasurementResponse
		err = json.Unmarshal(response.Data, &data)

		if err != nil {
			return errors.Wrap(err, "json.Unmarshal")
		}

		ch <- data.LiveMeasurement
	}
}
