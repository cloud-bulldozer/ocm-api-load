package helpers

import (
	"context"
	"net/http"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	sdk "github.com/openshift-online/ocm-sdk-go"
)

// BuildConnection build the vegeta connection
// that is going to be used for testing
func BuildConnection(ctx context.Context, gateway, clientID, clientSecret, token string, logger logging.Logger) (*sdk.Connection, error) {
	conn, err := sdk.NewConnectionBuilder().
		Insecure(true).
		URL(gateway).
		Client(clientID, clientSecret).
		Tokens(token).
		Logger(logger).
		TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
			return &CleanTestTransport{Wrapped: wrapped, Logger: logger}
		}).
		BuildContext(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
