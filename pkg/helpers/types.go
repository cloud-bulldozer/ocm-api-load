package helpers

import (
	"time"

	sdk "github.com/openshift-online/ocm-sdk-go"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// TestOptions allows defining a test and all related test infrastructure in a
// way that can easily be executed by generic testing functions or providing
// a custom "handler" in the case of complex scenarios. It also eliminates
// sending a ton of arguments to each function.
type TestOptions struct {

	// The Test
	TestName string // name of the test. e.g. get-access-token
	Path     string // path of the endpoint. e.g. /api/v1/foo
	Method   string // HTTP Method
	Body     []byte // Only really used by generic test handlers
	Rate     vegeta.Rate
	Duration time.Duration

	// Test "Infrastructure"
	ID         string                         // Unique UUID of a given test-suite execution.
	Handler    func(*TestOptions) (err error) // Function which tests the given endpoint
	Attacker   *vegeta.Attacker
	Connection *sdk.Connection
	Encoder    *vegeta.Encoder // Encodes results and writes them to a File
}
