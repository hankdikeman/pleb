/*
 * Directory APIs.
 */

package main

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"context"
	"os"
	"syscall"
)

// directory structure. matches interface fs.Node
type Dir struct{}

// for the given directory, return the directory attributes
func (dir Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 1
	attr.Mode = os.ModeDir | 0o555
	return nil
}

// for the given directory, lookup a file by name
func (dir Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	// TODO just a small testfile for now
	if name == "plebtest" {
		return File{}, nil
	}
	return nil, syscall.ENOENT
}

// for the given directory, return all their entries
func (dir Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Inode: 2, Name: "plebtest", Type: fuse.DT_File},
	}, nil
}
