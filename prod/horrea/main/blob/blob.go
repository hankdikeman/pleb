/*
 * Interface for interacting with data blobs.
 */

package blob

import (
	pb "github.com/pleb/prod/horrea/pb"

	"fmt"
	"log"
)

type Blob struct {
	content  []byte
	readOnly bool
	info     *pb.BlobInfo
}

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

// read the content specified by the Blob into the buffer
func (blob *Blob) ReadContent() error {
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

// Print summary information about the blob.
func (blob *Blob) ToString() string {
	return fmt.Sprintf("%s:%s, %s, %d bytes", blob.info.Major,
		blob.info.Minor, blob.info.BlobType.String(), blob.info.Size)
}
