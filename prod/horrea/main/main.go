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

	"github.com/pleb/prod/common/config"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/signal"
	"syscall"
)

type HorreaConfig struct {
	Port           int    `env:"PORT"         envDefault:55412`
	ChunkSizeKiB   int    `env:"CSIZEKIB"    envDefault:"64"`
	MaxFileSizeGiB int    `env:"FSIZEGIB"    envDefault:"10"`
	LocalBacked    bool   `env:"LOCALBACKED"  envDefault:"false"`
	LocalDirectory string `env:"LOCALDIR"  envDefault:"/tmp/pleb"`
}

const cfgPrefix = "HORREA_"

var cfg = HorreaConfig{}

type server struct {
	pb.UnimplementedHorreaServer
}

// shared PUT API, with streamed input and type specified by request.
func (s *server) PutContent(stream pb.Horrea_PutContentServer) error {
	// initial message provides message attributes
	in, err := stream.Recv()
	if err != nil {
		return err
	}
	info := in.GetInfo()

	// create buffer of specified size.
	writeBlob := blob.CreateBlob(info)

	// pass stream to type-specific PUT
	log.Printf("Request to put BLOB %s", writeBlob.ToString())

	for {
		// receive additional content from client stream
		in, err = stream.Recv()
		if err == io.EOF {
			break // client done sending
		} else if err != nil {
			log.Printf("Error receiving from client, %v", err)
			return err
		}
		// append content to internal buffer
		err = writeBlob.AppendChunk(in.GetChunk().Data)
		if err != nil {
			panic(err) // should never overfill buffer
		}
	}

	// file we write cannot be empty (for now)
	if writeBlob.GetSize() != writeBlob.GetCapacity() {
		fmt.Errorf("expected file %d bytes, is %d bytes",
			writeBlob.GetCapacity(), writeBlob.GetSize())
	}

	// push Blob to persistent storage
	err = writeBlob.WriteContent()
	if err != nil {
		log.Printf("Unable to persist received content, %v", err)
		return err
	}
	stream.SendAndClose(&empty.Empty{})

	return nil
}

// shared GET API, with streamed output and type-specified request.
func (s *server) GetContent(in *pb.GetContentReq, stream pb.Horrea_GetContentServer) error {
	// create buffer of specified size.
	readBlob := blob.CreateBlob(in.Info)
	readBlob.SetReadOnly() // GET is readonly

	// pass stream to type-specific GET
	log.Printf("Request to get BLOB %s", readBlob.ToString())
	err := readBlob.ReadContent()
	if err != nil {
		return err
	}

	// push retrieved data to output
	iter := blob.CreateBlobIterator(cfg.ChunkSizeKiB, 0)
	for {
		// pop additional chunk from blob
		dataout, err := readBlob.PopChunk(iter)
		if err != nil {
			break
		}
		// send next data chunk to client
		if err := stream.Send(&pb.Chunk{Data: dataout}); err != nil {
			return err
		}
	}

	return nil
}

// entrypoint for horrea server.
func main() {
	config.LoadConfig(&cfg, cfgPrefix)

	// watch for shutdown signals (XXX) needs to be in common package
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// additional initialization based on config
	err := blob.ConfigureBackend(cfg.LocalBacked, cfg.LocalDirectory)
	if err != nil {
		log.Fatalf("failed to init blob backend: %v", err)
	}

	// Start listening on server port
	log.Printf("Starting horrea server on port %d", cfg.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start gRPC server
	s := grpc.NewServer()
	go func() {
		pb.RegisterHorreaServer(s, &server{})
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// block on program exit
	<-ctx.Done()
	log.Printf("shutting down horrea server")
	s.GracefulStop()
}
