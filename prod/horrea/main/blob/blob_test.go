/*
 * Tests for Blob package. Encapsulates file IO operations.
 */

package blob

import (
	pb "github.com/pleb/prod/horrea/pb"

	"fmt"
	"testing"
)

// utility function to check that two byte arrays are equal
func checkBytesEqual(b1, b2 []byte) error {
	if len(b1) != len(b2) {
		return fmt.Errorf("byte arrays different sizes, %d != %d", len(b1), len(b2))
	}
	for i := range b1 {
		if b1[i] != b2[i] {
			return fmt.Errorf("byte %d not equal, %d != %d", i, b1[i], b2[i])
		}
	}
	return nil
}

// test that Blobs are created with correct attributes from Info
func TestBlobCreate(t *testing.T) {
	size := 1024 * 1024
	blobInfo := &pb.BlobInfo{
		Size: int64(size),
	}

	// internal buffer creation tests
	testblob := CreateBlob(blobInfo)
	if testblob.GetCapacity() != size {
		t.Fatalf("blob created with capacity %d, expected %d",
			testblob.GetCapacity(), size)
	}

	// read-only tests
	if testblob.readOnly {
		t.Fatalf("blob defaulted to ReadOnly")
	}
	testblob.SetReadOnly()
	if !testblob.readOnly {
		t.Fatalf("blob was not correctly set ReadOnly")
	}
}

// test Blob buffer manipulation functions
func TestBlobBufferIO(t *testing.T) {
	size := 1024 * 1024
	blobInfo := &pb.BlobInfo{
		Size: int64(size),
	}

	// create empty blob object
	testblob := CreateBlob(blobInfo)
	if testblob.GetCapacity() != size {
		t.Fatalf("capacity %d != %d", testblob.GetCapacity(), size)
	} else if testblob.GetSize() != 0 {
		t.Fatalf("blob size %d, should start 0", testblob.GetSize())
	}

	// create some testdata where byte_i == i % BYTE_MAX
	testdata := make([]byte, size, size)
	for b := range testdata {
		testdata[b] = byte(67)
	}

	// append testdata to the blob, should be halfway full
	err := testblob.AppendChunk(testdata)
	if err != nil {
		t.Fatalf("blob append returned error: %s", err)
	} else if testblob.GetCapacity() != size {
		t.Fatalf("capacity %d != %d after append", testblob.GetCapacity(), size)
	} else if testblob.GetSize() != size {
		t.Fatalf("size %d != %d after append", testblob.GetCapacity(), size)
	}

	// read testdata directly from the blob and compare contents
	if err = checkBytesEqual(testblob.content, testdata); err != nil {
		t.Fatalf("testdata and blob buffer content not equal after append, %d", err)
	}

	// use the chunk iterator to do the same, with different chunk sizes
	chunksizes := [5]int{23, 1024, 55 * 1024, 64 * 1024, 1024 * 1024 * 1024}
	for _, chunksize := range chunksizes {
		iter := CreateBlobIterator(chunksize, 0)
		for {
			// determine which testdata chunk should get grabbed
			start, end := iter.next, iter.next+iter.chunkSize
			if end > size {
				end = size
			}
			expectchunk := testdata[start:end]

			// read the chunk from the blob using iterator
			chunk, err := testblob.PopChunk(iter)
			if err != nil {
				break
			}
			if err = checkBytesEqual(chunk, expectchunk); err != nil {
				t.Fatalf("popped chunk != corresponding data, [%d, %d], %s",
					start, end, err)
			}
		}
		// check that the iterator actually read the whole blob
		if iter.next != size {
			t.Fatalf("iterator did not read full buffer, read %d/%d bytes",
				iter.next, size)
		}
	}

}
