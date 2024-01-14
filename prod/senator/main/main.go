/*
 * Remote filesystem handler.
 */

package main

import (
	pb "github.com/pleb/prod/senator/pb"

	"github.com/pleb/prod/common/bootstrap"
	"github.com/pleb/prod/common/config"

	"google.golang.org/grpc"

	"context"
	"fmt"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedSenatorServer
}

type SenatorConfig struct {
	Port int `env:"PORT"         envDefault:55417`
}

const cfgPrefix = "SENATOR_"

var (
	cfg = SenatorConfig{}
	srv = grpc.NewServer()
)

// load configuration
func setup() error {
	config.LoadConfig(&cfg, cfgPrefix)
	return nil
}

// run gRPC server and wait for shutdown
func run(done context.CancelFunc) {
	defer done()
	log.Printf("Starting senator server on port %d", cfg.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}
	defer lis.Close()
	pb.RegisterSenatorServer(srv, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
		return
	}
}

// gracefully stop server
func shutdown() {
	log.Printf("shutting down senator server")
	srv.GracefulStop()
}

// entrypoint for senator server.
func main() {
	bootstrap.RunDaemon(setup, run, shutdown)
}
