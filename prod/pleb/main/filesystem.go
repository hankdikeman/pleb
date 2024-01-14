/*
 * Filesystem APIs.
 */

package main

import (
	"bazil.org/fuse/fs"
)

// filesystem structure. matches interface fs.FS
type FS struct{}

// for the given filesystem, return the root directory
func (filesys FS) Root() (fs.Node, error) {
	return Dir{}, nil
}
