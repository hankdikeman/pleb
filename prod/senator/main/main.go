/*
 * Remote filesystem handler.
 */

package main

import (
	pb "github.com/pleb/prod/senator/pb"

	"github.com/caarlos0/env/v10"
	"google.golang.org/grpc"

	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
)

type SenatorConfig struct {
	Port int `env:"S_PORT"         envDefault:55417`
}

var config = SenatorConfig{}

type server struct {
	pb.UnimplementedSenatorServer
}

// entrypoint for senator server.
func main() {
	// Load config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("could not parse environment config: %v", err)
	}
	log.Printf("%+v\n", config)

	// watch for shutdown signals (XXX) needs to be in common package
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// Start listening on server port
	log.Printf("Starting senator server on port %d", config.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start gRPC server
	s := grpc.NewServer()
	go func() {
		pb.RegisterSenatorServer(s, &server{})
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// block on program exit
	<-ctx.Done()
	log.Printf("shutting down senator server")
	s.GracefulStop()
}
