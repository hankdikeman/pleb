/*
 * Entrypoint for AWS S3 frontend.
 */

/*
 * Serves as a frontend to manage access to S3 object storage
 * for data not suitable for storage in typical KV store.
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

// shared PUT API, with streamed input and type specified by request.
func (s *server) PutContent(stream pb.Horrea_PutContentServer) error {
	log.Printf("Request to put content.")
	return nil
}

// shared GET API, with streamed output and type-specified request.
func (s *server) GetContent(in *pb.GetContentReq, stream pb.Horrea_GetContentServer) error {
	log.Printf("Request to stream content for req %v", in)
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
