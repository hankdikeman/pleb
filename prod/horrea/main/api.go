/*
 * Storage frontend server APIs.
 */

package main

import (
	"github.com/pleb/prod/horrea/main/blob"
	pb "github.com/pleb/prod/horrea/pb"

	"github.com/golang/protobuf/ptypes/empty"

	"fmt"
	"io"
	"log"
)

// shared PUT API, with streamed input and type specified by request.
func (srv *server) PutContent(stream pb.Horrea_PutContentServer) error {
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
func (srv *server) GetContent(in *pb.GetContentReq, stream pb.Horrea_GetContentServer) error {
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
