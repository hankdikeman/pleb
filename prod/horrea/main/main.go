/*
 * Entrypoint for AWS S3 frontend.
 */

/*
 * Serves as a frontend to manage access to S3 object storage
 * for data not suitable for storage in typical KV store.
 * Can also be configured in file mode, where data is stored
 * as files on the local machine rather than at a third-party.
 *
 * Consumers:
 *    Pleb      - reads inputs, pushes outputs
 *    Caesar    - pushes job inputs, manages outputs
 *
 * Consumes:
 *    Fabricae  - KV store, maps S3 storage -> identifiers
 *
 * TODO eventually should maintain a caching layer, but this
 * requires some work to shard inputs and will add state to
 * this service. But probably worth it for shared inputs.
 */

package main

import (
	"github.com/pleb/prod/horrea/main/blob"
	pb "github.com/pleb/prod/horrea/pb"

	"google.golang.org/grpc"

	"fmt"
	"io"
	"log"
	"net"
)

// TODO below should be configurable (at just boot maybe)
const CHUNKSIZE = 64 * 1024             // 64 KiB
const MAXSIZE = 1024 * 1024 * 1024 * 10 // 10 GiB
const PORT = 55412

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
		return err
	}

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
	iter := blob.CreateBlobIterator(CHUNKSIZE, 0)
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
	log.Printf("Starting horrea server on port 55412")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHorreaServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
