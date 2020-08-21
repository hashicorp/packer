package driver

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/vmware/govmomi/vapi/library"
)

type Library struct {
	driver  *Driver
	library *library.Library
}

func (d *Driver) FindContentLibraryByName(name string) (*Library, error) {
	lm := library.NewManager(d.restClient.client)
	l, err := lm.GetLibraryByName(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Library{
		library: l,
		driver:  d,
	}, nil
}

func (d *Driver) FindContentLibraryItem(libraryId string, name string) (*library.Item, error) {
	lm := library.NewManager(d.restClient.client)
	items, err := lm.GetLibraryItems(d.ctx, libraryId)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, nil
}

func (d *Driver) FindContentLibraryFileDatastorePath(isoPath string) string {
	log.Printf("Check if ISO path is a Content Library path")
	err := d.restClient.Login(d.ctx)
	if err != nil {
		log.Printf("vCenter client not available. ISO path not identified as a Content Library path")
		return isoPath
	}

	trimmedPath := strings.TrimLeft(isoPath, " ")
	trimmedPath = strings.TrimLeft(trimmedPath, "/")
	libraryName := strings.Split(trimmedPath, "/")[0]
	itemName := strings.Split(trimmedPath, "/")[1]
	isoFile := strings.Split(trimmedPath, "/")[2]

	lib, err := d.FindContentLibraryByName(libraryName)
	if err != nil {
		log.Printf("ISO path not identified as a Content Library path")
		return isoPath
	}
	log.Printf("ISO path identified as a Content Library path")
	log.Printf("Finding the equivalent datastore path for the Content Library ISO file path")
	libItem, err := d.FindContentLibraryItem(lib.library.ID, itemName)
	if err != nil {
		log.Printf("[WARN] Couldn't find item %s: %s", itemName, err.Error())
		log.Printf("Trying to use %s as the datastore path", isoPath)
		return isoPath
	}
	if libItem == nil {
		log.Printf("[WARN] Couldn't find item %s", itemName)
		log.Printf("Trying to use %s as the datastore path", isoPath)
		return isoPath
	}
	datastoreName := d.GetDatastoreName(lib.library.Storage[0].DatastoreID)
	libItemDir := fmt.Sprintf("[%s] contentlib-%s/%s", datastoreName, lib.library.ID, libItem.ID)

	isoFilePath, err := d.GetDatastoreFilePath(lib.library.Storage[0].DatastoreID, libItemDir, isoFile)
	if err != nil {
		log.Printf("[WARN] Couldn't find datastore ID path for %s", isoFile)
		log.Printf("Trying to use %s as the datastore path", isoPath)
		return isoPath
	}

	_ = d.restClient.Logout(d.ctx)
	return path.Join(libItemDir, isoFilePath)
}
