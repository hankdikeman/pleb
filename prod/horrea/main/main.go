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

	"log"
	"net"
)

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

	// TODO read blob content in chunks from client

	// TODO push blob content to persistent storage
	err = writeBlob.WriteContent()
	if err != nil {
		return nil
	}

	return nil
}

// shared GET API, with streamed output and type-specified request.
func (s *server) GetContent(in *pb.GetContentReq, stream pb.Horrea_GetContentServer) error {
	// create buffer of specified size. TODO wrap in struct
	// data := make([]byte, 0, in.Info.Size)
	readBlob := blob.CreateBlob(in.Info)
	readBlob.SetReadOnly() // GET is readonly

	// pass stream to type-specific GET
	log.Printf("Request to get BLOB %s", readBlob.ToString())
	err := readBlob.ReadContent()
	if err != nil {
		return err
	}

	// TODO push the retrieved content to output

	return nil
}

// entrypoint for horrea server.
func main() {
	log.Printf("Starting horrea server on port 55412")
	lis, err := net.Listen("tcp", ":55412")
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
