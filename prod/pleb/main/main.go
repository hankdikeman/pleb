/*
 * Filesystem client main.
 */

package main

import (
	"github.com/pleb/prod/common/bootstrap"
	"github.com/pleb/prod/common/config"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"context"
	"log"
)

type PlebConfig struct {
	MountPoint string `env:"MOUNTPOINT"  envDefault:"/mnt/pleb"`
}

const cfgPrefix = "PLEB_"

var cfg = PlebConfig{}

// TODO naked global is a bad idea
var conn *fuse.Conn

// mount filesystem, make remote filesystem connection
func setup() error {
	var err error
	config.LoadConfig(&cfg, cfgPrefix)

	/* TODO server auth and mount cleanup */

	conn, err = fuse.Mount(
		cfg.MountPoint,
		fuse.FSName("pleb"),
		fuse.Subtype("plebfs"),
	)
	if err != nil {
		log.Printf("Could not mount %s: %v",
			cfg.MountPoint, err)
	}
	return err
}

// serve the filesystem at the mountpoint
func run(done context.CancelFunc) {
	defer done()
	log.Printf("serving FS at %s", cfg.MountPoint)
	err := fs.Serve(conn, FS{})
	if err != nil {
		log.Printf("error serving filesystem: %v", err)
	}
}

// unmount the filesystem, close filesystem connection
func shutdown() {
	log.Printf("shutting down, unmounting %s", cfg.MountPoint)
	if err := fuse.Unmount(cfg.MountPoint); err != nil {
		/*
		 * (XXX) this is not good enough. with the
		 * mountpoint active, conn.Close() will hang
		 */
		log.Printf("could not unmount mountpoint %s: %v",
			cfg.MountPoint, err)
	}
	if err := conn.Close(); err != nil {
		log.Printf("could not close FUSE connection: %v", err)
	}
	log.Printf("closed connection, exiting")
}

// main method. mounts and starts serving on FUSE filesystem
func main() {
	bootstrap.RunDaemon(setup, run, shutdown)
}
