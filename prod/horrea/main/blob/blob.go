/*
 * Interface for interacting with data blobs.
 */

package blob

import (
	pb "github.com/pleb/prod/horrea/pb"

	"errors"
	"fmt"
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
var ErrNoExist = errors.New("cannot read, file doesn't exist")
var ErrNotSupp = errors.New("I/O mode not supported")

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
	var err error
	major, minor := blob.info.Major, blob.info.Minor
	blobtype := blob.info.BlobType.String()

	// TODO the file version is simple since all files are
	//  local and equivalent. The S3-backed version will need
	//  a subsidiary call to map major/minor to a real ID.
	blob.content, err = blobReadInternal(major, minor, blobtype)
	return err
}

// write the content currently contained in the Blob buffer
func (blob *Blob) WriteContent() error {
	major, minor := blob.info.Major, blob.info.Minor
	blobtype := blob.info.BlobType.String()

	// ReadOnly blobs not allowed to write
	if blob.readOnly {
		return ErrReadOnly
	}
	return blobWriteInternal(major, minor, blobtype, blob.content)
}

// return the underlying buffer of the Blob.
func (blob *Blob) GetBuffer() []byte {
	return blob.content
}

// return the BlobInfo used to create the blob
func (blob *Blob) GetBlobInfo() *pb.BlobInfo {
	return blob.info
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
