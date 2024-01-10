/*
 * Client wrapper object for remote filesystem access.
 */

package netclient

import (
	senator "github.com/pleb/prod/senator/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"context"
)

type NetClient struct {
	conn   *grpc.ClientConn
	client senator.SenatorClient
}

// create a net client object
func CreateNetClient(ctx context.Context, endpoint string) (*NetClient, error) {
	// dial remote filesystem connection and create client
	conn, err := grpc.DialContext(ctx,
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	// pack into NetClient and return
	netclient := new(NetClient)
	netclient.conn = conn
	netclient.client = senator.NewSenatorClient(conn)
	return netclient, nil
}
