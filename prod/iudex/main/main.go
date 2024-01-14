/*
 * Concurrency manager.
 */

package main

import (
	pb "github.com/pleb/prod/iudex/pb"

	"github.com/pleb/prod/common/bootstrap"
	"github.com/pleb/prod/common/config"

	"google.golang.org/grpc"

	"context"
	"fmt"
	"log"
	"net"
)

type IudexConfig struct {
	Port int `env:"PORT"         envDefault:55414`
}

const cfgPrefix = "IUDEX_"

var (
	cfg = IudexConfig{}
	srv = grpc.NewServer()
)

type server struct {
	pb.UnimplementedIudexServer
}

// load configuration
func setup() error {
	config.LoadConfig(&cfg, cfgPrefix)
	return nil
}

// run gRPC server and wait for shutdown
func run(done context.CancelFunc) {
	defer done()
	log.Printf("Starting iudex server on port %d", cfg.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}
	defer lis.Close()
	pb.RegisterIudexServer(srv, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
		return
	}
}

// gracefully stop server
func shutdown() {
	log.Printf("shutting down iudex server")
	srv.GracefulStop()
}

// entrypoint for iudex server.
func main() {
	bootstrap.RunDaemon(setup, run, shutdown)
}
