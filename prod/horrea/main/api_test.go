/*
 * Desktop tests for storage frontend server
 */

package main

import (
	"github.com/pleb/prod/horrea/main/blob"
	pb "github.com/pleb/prod/horrea/pb"

	"github.com/pleb/prod/common/config"

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

// setup the unit test. unit tests hardset to local mode
func setupTest() error {
	config.LoadConfig(&cfg, cfgPrefix)
	err := blob.ConfigureBackend(true, "/tmp/horrea-test")
	if err != nil {
		log.Fatalf("failed to configure backend: %v", err)
	}
	return err
}

// returns closure function to run server with given listener
func runServer(listener *bufconn.Listener) {
            // start test server on provided listener
		srv = grpc.NewServer()
		pb.RegisterHorreaServer(srv, &server{})
		if err := srv.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
}

// returns closure function to shutdown server on given listener
func createCloserFunc(listener *bufconn.Listener) func() {
	return func() {
		if err := listener.Close(); err != nil {
			log.Fatalf("error closing listener: %v", err)
		}
		srv.Stop()
	}
}

// start a test server and return a client + server stop function
func startTestServer(ctx context.Context) (pb.HorreaClient, func()) {
        // setup backend and config for test
        if err := setupTest() ; err != nil {
            log.Fatalf("could not setup test: %v", err)
        }

        // run server against buffered network listener
	bufsize := 10 * 1024 * 1024
        listener := bufconn.Listen(bufsize)
        go runServer(listener)

	// dial local buffered network context
	conn, err := grpc.DialContext(ctx,
		"",
		grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) {
				return listener.Dial()
			}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}
	client := pb.NewHorreaClient(conn)

	return client, createCloserFunc(listener)
}

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
