package handlers

import (
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestStaticEndpoint(options *types.TestOptions) error {

	// Specify the HTTP request(s) that will be executed
	target := vegeta.Target{
		Method: options.Method,
		URL:    options.Path,
	}
	if len(options.Body) > 0 {
		target.Body = options.Body
	}
	targeter := vegeta.NewStaticTargeter(target)

	// Execute the HTTP Requests; repeating as needed to meet the specified duration
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, options.TestName) {
		options.Encoder.Encode(res)
	}

	return nil

}
