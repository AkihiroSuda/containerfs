package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	netctx "golang.org/x/net/context"
)

type ContainerFs struct {
	pathfs.FileSystem
	mu         sync.Mutex
	client     *client.Client
	containers map[string]types.ContainerJSON
	idByName   map[string]string // key: name, value: id
}

// update updates fs.{containers, idByName}.
// todo: optimize
func (fs *ContainerFs) update() error {
	clist, err := fs.client.ContainerList(netctx.Background(),
		types.ContainerListOptions{})
	if err != nil {
		return err
	}
	idByName := make(map[string]string, 0) // key: name, value: id
	containers := make(map[string]types.ContainerJSON, 0)
	for _, c := range clist {
		cJSON, err := fs.client.ContainerInspect(netctx.Background(), c.ID)
		if err != nil {
			return err
		}
		containers[c.ID] = cJSON
		for _, cname := range c.Names {
			baseCName := filepath.Base(cname)
			idByName[baseCName] = c.ID
		}
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.containers = containers
	fs.idByName = idByName
	return nil
}

func (fs *ContainerFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	log.Printf("ENTER GetAttr(%s)", name)
	defer log.Printf("LEAVE GetAttr(%s)", name)

	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	for cid := range fs.containers {
		if cid == name {
			return &fuse.Attr{
				Mode: fuse.S_IFLNK | 0644,
			}, fuse.OK
		}
	}
	for cname := range fs.idByName {
		if cname == name {
			return &fuse.Attr{
				Mode: fuse.S_IFLNK | 0644,
			}, fuse.OK
		}
	}
	return nil, fuse.ENOENT
}

func (fs *ContainerFs) OnMount(nodeFs *pathfs.PathNodeFs) {
	log.Printf("ENTER OnMount()")
	defer log.Printf("LEAVE OnMount()")

	var err error
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost != "" {
		log.Printf("WARNING: DOCKER_HOST is set (%s). Make sure this is localhost.", dockerHost)
	}
	fs.client, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	info, err := fs.client.Info(netctx.Background())
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to Docker daemon: %+v", info)
	if err = fs.update(); err != nil {
		panic(err)
	}
}

func (fs *ContainerFs) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	log.Printf("ENTER OpenDir(%s)", name)
	defer log.Printf("LEAVE OpenDir(%s)", name)

	if name == "" {
		if err := fs.update(); err != nil {
			log.Printf("error: %v", err)
			return nil, fuse.EIO
		}
		var entries []fuse.DirEntry
		fs.mu.Lock()
		defer fs.mu.Unlock()
		for cid := range fs.containers {
			entry := fuse.DirEntry{Name: cid, Mode: fuse.S_IFLNK}
			entries = append(entries, entry)
		}
		for cname := range fs.idByName {
			entry := fuse.DirEntry{Name: cname, Mode: fuse.S_IFLNK}
			entries = append(entries, entry)
		}
		return entries, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (fs *ContainerFs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	log.Printf("ENTER Readlink(%s)", name)
	defer log.Printf("LEAVE Readlink(%s)", name)
	fs.mu.Lock()
	defer fs.mu.Unlock()
	for cid, c := range fs.containers {
		if cid == name {
			proc := filepath.Join("/proc", strconv.Itoa(c.State.Pid))
			root := filepath.Join(proc, "root")
			return root, fuse.OK
		}
	}
	for cname, cid := range fs.idByName {
		if cname == name {
			return cid, fuse.OK
		}
	}
	return "", fuse.ENOENT
}
