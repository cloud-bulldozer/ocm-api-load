package ocm

import (
	"context"
	"fmt"
	"net/http"

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
	conn, err := sdk.NewConnectionBuilder().
		Insecure(true).
		URL(gateway).
		Client(clientID, clientSecret).
		Tokens(token).
		Logger(logger).
		TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
			return &CleanClustersTransport{Wrapped: wrapped, Logger: logger, Context: ctx}
		}).
		BuildContext(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func BuildConnections(ctx context.Context, logger logging.Logger) ([]interface{}, error) {
	connections := make([]interface{}, 0)
	auths := viper.GetStringMap("ocm")["auths"].([]interface{})
	for i, a := range auths {
		m := a.(map[interface{}]interface{})
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
		defer Cleanup(connection)

		connections = append(connections, Connection{
			Name:       fmt.Sprintf("%d", i),
			Logger:     logger,
			Ctx:        ctx,
			Connection: connection})
	}
	return connections, nil
}
