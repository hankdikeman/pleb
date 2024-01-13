/*
 * Storage frontend server.
 */

package main

import (
	pb "github.com/pleb/prod/fabricae/pb"

	"github.com/pleb/prod/common/config"

	"google.golang.org/grpc"

	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
)

type FabricaeConfig struct {
	Port int `env:"PORT"         envDefault:55416`
}

const cfgPrefix = "FABRICAE_"

var cfg = FabricaeConfig{}

type server struct {
	pb.UnimplementedFabricaeServer
}

// entrypoint for fabricae server.
func main() {
	config.LoadConfig(&cfg, cfgPrefix)

	// watch for shutdown signals (XXX) needs to be in common package
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Start listening on server port
	log.Printf("Starting fabricae server on port %d", cfg.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start gRPC server
	s := grpc.NewServer()
	go func() {
		pb.RegisterFabricaeServer(s, &server{})
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// block on program exit
	<-ctx.Done()
	log.Printf("shutting down fabricae server")
	s.GracefulStop()
}
