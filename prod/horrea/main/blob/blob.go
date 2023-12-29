/*
 * Interface for interacting with data blobs.
 */

package blob

import (
	pb "github.com/pleb/prod/horrea/pb"

	"errors"
	"fmt"
	"log"
)

type Blob struct {
	content  []byte       // blob content.
	readOnly bool         // read only?
	info     *pb.BlobInfo // Blob information.
}

type BlobIterator struct {
	chunkSize int // chunk size to read.
	next      int // next byte index.
}

// Blob-specific errors
var ErrEndOfBuffer = errors.New("passed end of buffer")
var ErrReadOnly = errors.New("cannot write, read-only blob")

// Blob constructor.
func CreateBlob(info *pb.BlobInfo) *Blob {
	blob := new(Blob)
	blob.content = make([]byte, 0, info.Size)
	blob.readOnly = false
	blob.info = info
	return blob
}

// Set ReadOnly to block underlying blob mutation.
func (blob *Blob) SetReadOnly() {
	blob.readOnly = true
}

// append a chunk to the Blob, return EOF if past max size.
func (blob *Blob) AppendChunk(data []byte) error {
	if cap(blob.content) < len(blob.content)+len(data) {
		return ErrEndOfBuffer
	}
	blob.content = append(blob.content, data...)
	return nil
}

func CreateBlobIterator(size, start int) *BlobIterator {
	iter := new(BlobIterator)
	iter.chunkSize = size
	iter.next = start
	return iter
}

// pop a chunk from the buffer, return EOF if iterator reaches end
func (blob *Blob) PopChunk(iter *BlobIterator) ([]byte, error) {
	start, end := iter.next, iter.next+iter.chunkSize
	if len(blob.content) <= start {
		return nil, ErrEndOfBuffer
	} else if len(blob.content) <= end {
		end = len(blob.content)
	}
	iter.next = end
	return blob.content[start:end], nil
}

// read the content specified by the Blob into the buffer
func (blob *Blob) ReadContent() error {
	// TODO fill in file version at least
	switch blobType := blob.info.BlobType; blobType {
	case pb.BlobType_Raw:
		log.Print("No-op RAW read")
	case pb.BlobType_Tool:
		log.Print("No-op TOOL read")
	case pb.BlobType_Input:
		log.Print("No-op INPUT read")
	case pb.BlobType_Output:
		log.Print("No-op OUTPUT read")
	}
	return nil
}

// write the content currently contained in the Blob buffer
func (blob *Blob) WriteContent() error {
	if blob.readOnly {
		return ErrReadOnly
	}

	// TODO fill in file version at least
	switch blobType := blob.info.BlobType; blobType {
	case pb.BlobType_Raw:
		log.Print("No-op RAW write")
	case pb.BlobType_Tool:
		log.Print("No-op TOOL write")
	case pb.BlobType_Input:
		log.Print("No-op INPUT write")
	case pb.BlobType_Output:
		log.Print("No-op OUTPUT write")
	}
	return nil
}

// return the underlying buffer of the Blob.
func (blob *Blob) GetBuffer() []byte {
	return blob.content
}

// return total capacity of underlying buffer
func (blob *Blob) GetCapacity() int {
	return cap(blob.content)
}

// return current size of underlying buffer
func (blob *Blob) GetSize() int {
	return len(blob.content)
}

// Print summary information about the blob.
func (blob *Blob) ToString() string {
	return fmt.Sprintf("%s:%s, %s, %d bytes", blob.info.Major,
		blob.info.Minor, blob.info.BlobType.String(), blob.info.Size)
}
