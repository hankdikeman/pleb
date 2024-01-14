/*
 * Entrypoint for storage frontend server.
 */

/*
 * Serves as a frontend for bulk data storage (i.e., everything
 * which is not file metadata). Can operate in local file mode
 * or cloud storage mode.
 *
 * Consumers:
 *    Senator   - Serves file requests, including Reads/Writes
 *
 * Consumes:
 *    Iudex     - Guards concurrent file accesses
 *
 * TODO eventually should maintain a caching layer, but this
 * requires some work to shard inputs and will add state to
 * this service. But probably worth it for shared inputs.
 * TODO the cloud storage aspect is unfinished right now.
 */

package main

import (
	"github.com/pleb/prod/horrea/main/blob"
	pb "github.com/pleb/prod/horrea/pb"

	"github.com/pleb/prod/common/bootstrap"
	"github.com/pleb/prod/common/config"

	"google.golang.org/grpc"

	"context"
	"fmt"
	"log"
	"net"
)

type HorreaConfig struct {
	Port           int    `env:"PORT"         envDefault:55412`
	ChunkSizeKiB   int    `env:"CSIZEKIB"    envDefault:"64"`
	MaxFileSizeGiB int    `env:"FSIZEGIB"    envDefault:"10"`
	LocalBacked    bool   `env:"LOCALBACKED"  envDefault:"false"`
	LocalDirectory string `env:"LOCALDIR"  envDefault:"/tmp/pleb"`
}

const cfgPrefix = "HORREA_"

var (
	cfg = HorreaConfig{}
	srv = grpc.NewServer()
)

type server struct {
	pb.UnimplementedHorreaServer
}

// load configuration and initialize backend
func setup() error {
	config.LoadConfig(&cfg, cfgPrefix)
	// additional initialization based on config
	err := blob.ConfigureBackend(cfg.LocalBacked,
		cfg.LocalDirectory)
	if err != nil {
		log.Printf("failed to init blob backend: %v", err)
	}
	return err
}

// run gRPC server and wait for shutdown
func run(done context.CancelFunc) {
	defer done()
	log.Printf("Starting horrea server on port %d", cfg.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return
	}
	defer lis.Close()
	pb.RegisterHorreaServer(srv, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
		return
	}
}

// gracefully stop server
func shutdown() {
	log.Printf("shutting down horrea server")
	srv.GracefulStop()
}

// entrypoint for horrea server.
func main() {
	bootstrap.RunDaemon(setup, run, shutdown)
}
