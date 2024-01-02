/*
 * Desktop tests for storage frontend server
 */

package main

import (
	"github.com/pleb/prod/horrea/main/blob"
	pb "github.com/pleb/prod/horrea/pb"

	"github.com/caarlos0/env/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"testing"
)

// make a data blob for test purposes
func makeTestBlob(size, major, minor int) *blob.Blob {
	// create pseudorandom data
	data := make([]byte, size, size)
	for b := range data {
		data[b] = byte(rand.Intn(256))
	}
	// create test blob and append test data
	testblob := blob.CreateBlob(&pb.BlobInfo{
		Size:     int64(size),
		Major:    fmt.Sprintf("%d", major),
		Minor:    fmt.Sprintf("%d", minor),
		BlobType: pb.BlobType_Raw,
	})
	err := testblob.AppendChunk(data)
	if err != nil {
		log.Fatalf("could not append chunk to blob, %v", err)
	}
	return testblob
}

// start a test server and return a client + server stop function
func startTestServer(ctx context.Context) (pb.HorreaClient, func()) {
	// create a local buffer to emulate network
	buffer := 1024 * 1024 * 10
	lis := bufconn.Listen(buffer)

	// do server configuration and backend setup
	if err := env.Parse(&config); err != nil {
		log.Fatalf("failed to parse environment config: %v", err)
	}
	err := blob.ConfigureBackend(true, "/tmp/horrea-test")
	if err != nil {
		log.Fatalf("failed to configure backend: %v", err)
	}

	// register gRPC test server
	s := grpc.NewServer()
	pb.RegisterHorreaServer(s, &server{})

	// start separate server thread
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
		}
	}()

	// dial local buffered network context
	conn, err := grpc.DialContext(ctx,
		"",
		grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) {
				return lis.Dial()
			}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}

	// create anonymous server shutdown function and client
	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		s.Stop()
	}
	client := pb.NewHorreaClient(conn)

	return client, closer
}

// just test server startup logic, but don't submit client calls
func TestHorreaServerStartup(t *testing.T) {
	ctx := context.Background()

	_, closer := startTestServer(ctx)
	defer closer()
}

// test basic server write and readback
func TestHorreaServerBasic(t *testing.T) {
	ctx := context.Background()

	client, closer := startTestServer(ctx)
	defer closer()

	// clients will have their own data iterators, but we
	// can use the internal blob package for internal testing
	size, chunksize := 1024*1024, 64*1024
	writeblob := makeTestBlob(size, rand.Int(), rand.Int())

	// create the client stream
	wstream, err := client.PutContent(ctx)
	if err != nil {
		t.Fatalf("could not create client context, %v", err)
	}
	// send the initial PUT request with metadata
	wstream.Send(&pb.PutContentReq{
		Input: &pb.PutContentReq_Info{
			Info: writeblob.GetBlobInfo(),
		},
	})

	// stream the data to the server
	iter := blob.CreateBlobIterator(chunksize, 0)
	for {
		chunk, err := writeblob.PopChunk(iter)
		if err != nil {
			break
		}
		err = wstream.Send(&pb.PutContentReq{
			Input: &pb.PutContentReq_Chunk{
				Chunk: &pb.Chunk{Data: chunk},
			},
		})
		if err != nil {
			t.Fatalf("could not send data chunk, %v", err)
		}
	}

	// close the stream and finalize PUT
	_, err = wstream.CloseAndRecv()
	if err != nil {
		t.Fatalf("data could not be persisted, %v", err)
	}

	// read data back from the server and compare
	rstream, err := client.GetContent(
		ctx,
		&pb.GetContentReq{
			Info: writeblob.GetBlobInfo(),
		},
	)
	if err != nil {
		t.Fatalf("could not create read client context, %v", err)
	}

	// create a readblob to hold the data and stream from server
	readblob := blob.CreateBlob(writeblob.GetBlobInfo())
	for {
		in, err := rstream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("error receiving from server, %v", err)
		}
		readblob.AppendChunk(in.Data)
	}

	// two buffers should be the same size and content
	readbuf, writebuf := readblob.GetBuffer(), writeblob.GetBuffer()
	if len(readbuf) != len(writebuf) {
		t.Fatalf("did not read all data back from server, %d != %d",
			len(readbuf), len(writebuf))
	}
	for i := range readbuf {
		if readbuf[i] != writebuf[i] {
			t.Fatalf("byte %d not read back correctly, %x != %x",
				i, readbuf[i], writebuf[i])
		}
	}
}
