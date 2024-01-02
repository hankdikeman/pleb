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

	"github.com/caarlos0/env/v10"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"fmt"
	"io"
	"log"
	"net"
)

type HorreaConfig struct {
	Port           int    `env:"H_PORT"         envDefault:55412`
	ChunkSizeKiB   int    `env:"H_CSIZEKIB"    envDefault:"64"`
	MaxFileSizeGiB int    `env:"H_FSIZEGIB"    envDefault:"10"`
	LocalBacked    bool   `env:"H_LOCALBACKED"  envDefault:"false"`
	LocalDirectory string `env:"H_LOCALDIR"  envDefault:"/tmp/pleb"`
}

var config = HorreaConfig{}

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
	iter := blob.CreateBlobIterator(config.ChunkSizeKiB, 0)
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
	// Load config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("could not parse environment config: %v", err)
	}
	log.Printf("%+v\n", config)

	// additional initialization based on config
	err := blob.ConfigureBackend(config.LocalBacked, config.LocalDirectory)
	if err != nil {
		log.Fatalf("failed to init blob backend: %v", err)
	}

	// Start listening on server port
	log.Printf("Starting horrea server on port %d", config.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start gRPC server
	s := grpc.NewServer()
	pb.RegisterHorreaServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
