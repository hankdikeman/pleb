/*
 * Low-level interface for persisting and reading blobs.
 */

package blob

import (
	"log"
	"os"
	"path/filepath"
)

type BlobWriter struct {
	initialized    bool   // was the IO interface initialized?
	localbacked    bool   // is the blob package running in local file mode?
	localdirectory string // which directory to use for storage
}

var blobio = &BlobWriter{}

func ConfigureBackend(localbacked bool, localdirectory string) error {
	// mark the storage IO as initialized
	blobio.initialized = true
	// configure local IO if required
	if localbacked {
		blobio.localbacked = true
		blobio.localdirectory = localdirectory

		// if local directory does not exist, create it
		log.Printf("Creating local file directory %s", localdirectory)
		err := os.MkdirAll(localdirectory, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// construct filepath from identifiers + blob
func constructFilePath(major, minor, blobtype string) string {
	filename := major + "." + minor + "." + blobtype
	return filepath.Join(blobio.localdirectory, filename)
}

// read a byte buffer from disk or cloud storage
func blobReadInternal(major, minor, blobtype string) ([]byte, error) {
	if !blobio.localbacked {
		return nil, ErrNotSupp
	}
	readpath := constructFilePath(major, minor, blobtype)
	log.Printf("Reading local file content from %s", readpath)
	return os.ReadFile(readpath)
}

// persist a byte buffer to disk or cloud storage
func blobWriteInternal(major, minor, blobtype string, content []byte) error {
	if !blobio.localbacked {
		return ErrNotSupp
	}
	writepath := constructFilePath(major, minor, blobtype)
	log.Printf("Writing local file content to %s", writepath)
	return os.WriteFile(writepath, content, 0644)
}
