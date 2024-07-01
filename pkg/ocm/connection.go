package ocm

import (
	"context"
	"net/http"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/viper"
)

type Connector interface {
	GetConnection() *sdk.Connection
}

type Connection struct {
	Name       string
	Logger     logging.Logger
	Connection *sdk.Connection
	Ctx        context.Context
}

func (c Connection) GetConnection() *sdk.Connection {
	c.Logger.Info(c.Ctx, "Connection ID: %s responding", c.Name)
	return c.Connection
}

// BuildConnection build the vegeta connection
// that is going to be used for testing
func BuildConnection(gateway, clientID, clientSecret, token string, logger logging.Logger, ctx context.Context) (*sdk.Connection, error) {
	if token == "" {
		conn, err := sdk.NewConnectionBuilder().
			Insecure(true).
			URL(gateway).
			Client(clientID, clientSecret).
			Logger(logger).
			TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
				return &helpers.CleanTestTransport{Wrapped: wrapped, Logger: logger}
			}).
			BuildContext(ctx)
		if err != nil {
			return nil, err
		}
		return conn, nil
	} else {
		conn, err := sdk.NewConnectionBuilder().
			Insecure(true).
			URL(gateway).
			Client(clientID, clientSecret).
			Tokens(token).
			Logger(logger).
			TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
				return &helpers.CleanTestTransport{Wrapped: wrapped, Logger: logger}
			}).
			BuildContext(ctx)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

}

func BuildConnections(ctx context.Context, logger logging.Logger) ([]*sdk.Connection, error) {
	connections := make([]*sdk.Connection, 0)

	var auths []interface{}
	if viper.Sub("ocm") != nil {
		auths = viper.GetStringMap("ocm")["auths"].([]interface{})
	} else if (viper.GetString("ocm-token") != "") || (viper.GetString("client-id") != "" && viper.GetString("client-secret") != "") {
		auth := map[string]interface{}{
			"token":         viper.GetString("ocm-token"),
			"client-id":     viper.GetString("client-id"),
			"client-secret": viper.GetString("client-secret"),
		}
		auths = append(auths, auth)
	}
	for _, a := range auths {
		m := a.(map[string]interface{})
		token, ok := m["token"]
		if !ok {
			token = ""
		}
		clientID, ok := m["client-id"]
		if !ok {
			clientID = ""
		}
		clientSecret, ok := m["client-secret"]
		if !ok {
			clientSecret = ""
		}
		connection, err := BuildConnection(viper.GetString("gateway-url"),
			clientID.(string),
			clientSecret.(string),
			token.(string),
			logger,
			ctx)
		if err != nil {
			logger.Fatal(ctx, "creating api connection: %v - %v", token, err)
		}
		defer helpers.Cleanup(ctx, connection)

		connections = append(connections, connection)
	}
	return connections, nil
}
