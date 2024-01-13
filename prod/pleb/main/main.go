/*
 * Filesystem client. Hooks local system to remote filesystem.
 */

package main

import (
	// "github.com/pleb/prod/pleb/main/netclient"

	"github.com/pleb/prod/common/bootstrap"
	"github.com/pleb/prod/common/config"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"

	"context"
	"log"
	"os"
	"syscall"
)

type PlebConfig struct {
	MountPoint string `env:"MOUNTPOINT"  envDefault:"/mnt/pleb"`
}

const cfgPrefix = "PLEB_"

var cfg = PlebConfig{}

// filesystem structure. matches interface fs.FS
type FS struct{}

// for the given filesystem, return the root directory
func (filesys FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

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

// main method. mounts and starts serving on FUSE filesystem
func main() {
	var conn *fuse.Conn

	bootstrap.RunDaemon(
		func() error {
			var err error
			config.LoadConfig(&cfg, cfgPrefix)
			conn, err = fuse.Mount(
				cfg.MountPoint,
				fuse.FSName("pleb"),
				fuse.Subtype("plebfs"),
			)
			if err != nil {
				log.Printf("Could not mount %s: %v", cfg.MountPoint, err)
			}
			return err
		},
		func(done context.CancelFunc) {
			defer done()
			log.Printf("serving FS at %s", cfg.MountPoint)
			err := fs.Serve(conn, FS{})
			if err != nil {
				log.Printf("error serving filesystem: %v", err)
			}
		},
		func() {
			log.Printf("shutting down, unmounting %s", cfg.MountPoint)
			err := fuse.Unmount(cfg.MountPoint)
			if err != nil {
				/*
				 * (XXX) this is not good enough. with the
				 * mountpoint active, conn.CLose() will hang
				 */
				log.Printf("could not unmount mountpoint %s: %v",
					cfg.MountPoint, err)
			}
			err = conn.Close()
			if err != nil {
				log.Printf("could not close FUSE connection: %v", err)
			}
			log.Printf("closed connection, exiting")
		},
	)
}
