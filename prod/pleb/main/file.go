/*
 * File APIs.
 */

package main

import (
	"bazil.org/fuse"

	"context"
)

// file structure. matches interface fs.Node
type File struct{}

// for the given file, return the file attrs
func (f File) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 2
	attr.Mode = 0o444
	attr.Size = uint64(len("hello plebs\n"))
	return nil
}

// for the given file, return the content as a byte array
func (f File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte("hello plebs\n"), nil
}
