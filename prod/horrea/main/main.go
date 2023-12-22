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
	"flag"
	"fmt"
	pb "github.com/pleb/prod/horrea/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedHorreaServer
}

var (
	port = flag.Int("port", 50444, "The server port")
)

func blobInfoToString(info *pb.BlobInfo) string {
	return fmt.Sprintf("%s:%s, %s, %d bytes", info.Major, info.Minor, info.BlobType.String(), info.Size)
}

// shared PUT API, with streamed input and type specified by request.
func (s *server) PutContent(stream pb.Horrea_PutContentServer) error {
	// initial message provides message attributes
	in, err := stream.Recv()
	if err != nil {
		return err
	}
	info := in.GetInfo()

	// create buffer of specified size. TODO wrap in struct
	// data := make([]byte, 0, info.Size)

	// pass stream to type-specific PUT
	log.Printf("Request to put BLOB %s", blobInfoToString(info))
	switch blobType := info.BlobType; blobType {
	case pb.BlobType_Raw:
		log.Print("No-op RAW put")
	case pb.BlobType_Tool:
		log.Print("No-op TOOL put")
	case pb.BlobType_Input:
		log.Print("No-op INPUT put")
	case pb.BlobType_Output:
		log.Print("No-op OUTPUT put")
	default:
		return fmt.Errorf("BlobType %d not supported", info.BlobType)
	}
	return nil
}

// shared GET API, with streamed output and type-specified request.
func (s *server) GetContent(in *pb.GetContentReq, stream pb.Horrea_GetContentServer) error {
	// create buffer of specified size. TODO wrap in struct
	// data := make([]byte, 0, in.Info.Size)

	// pass stream to type-specific GET
	log.Printf("Request to get BLOB %s", blobInfoToString(in.Info))
	switch blobType := in.Info.BlobType; blobType {
	case pb.BlobType_Raw:
		log.Print("No-op RAW get")
	case pb.BlobType_Tool:
		log.Print("No-op TOOL get")
	case pb.BlobType_Input:
		log.Print("No-op INPUT get")
	case pb.BlobType_Output:
		log.Print("No-op OUTPUT get")
	default:
		return fmt.Errorf("BlobType %d not supported", in.Info.BlobType)
	}
	return nil
}

// entrypoint for horrea server. Port specified as input argument.
func main() {
	flag.Parse()
	log.Printf("Starting horrea server on port %d", *port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
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
