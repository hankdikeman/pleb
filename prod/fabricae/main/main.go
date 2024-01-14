/*
 * Storage frontend server.
 */

package main

import (
	pb "github.com/pleb/prod/fabricae/pb"

	"github.com/pleb/prod/common/bootstrap"
	"github.com/pleb/prod/common/config"

	"google.golang.org/grpc"

	"context"
	"fmt"
	"log"
	"net"
)

type FabricaeConfig struct {
	Port int `env:"PORT"         envDefault:55416`
}

const cfgPrefix = "FABRICAE_"

var (
	cfg = FabricaeConfig{}
	srv = grpc.NewServer()
)

type server struct {
	pb.UnimplementedFabricaeServer
}

// load configuration
func setup() error {
	config.LoadConfig(&cfg, cfgPrefix)
	return nil
}

// run gRPC server and wait for shutdown
func run(done context.CancelFunc) {
	defer done()
	log.Printf("Starting fabricae server on port %d", cfg.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}
	defer lis.Close()
	pb.RegisterFabricaeServer(srv, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
		return
	}
}

// gracefully stop server
func shutdown() {
	log.Printf("shutting down fabricae server")
	srv.GracefulStop()
}

// entrypoint for fabricae server.
func main() {
	bootstrap.RunDaemon(setup, run, shutdown)
}
