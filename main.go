package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	mountpoint := flag.Arg(0)
	s, err := server(mountpoint, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	log.Printf("Serving at %s", mountpoint)
	log.Printf("FIXME: please run `%s -u %s` for unmounting", os.Args[0], mountpoint)
	s.Serve()
}

func server(mountpoint string, debug bool) (*fuse.Server, error) {
	fs := pathfs.NewPathNodeFs(&ContainerFs{FileSystem: pathfs.NewDefaultFileSystem()}, nil)
	fs.SetDebug(debug)
	server, _, err := nodefs.MountRoot(mountpoint, fs.Root(), nil)
	return server, err
}
