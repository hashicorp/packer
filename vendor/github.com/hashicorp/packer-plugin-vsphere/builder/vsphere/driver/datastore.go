package driver

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

type Datastore interface {
	Info(params ...string) (*mo.Datastore, error)
	FileExists(path string) bool
	DirExists(path string) bool
	Name() string
	ResolvePath(path string) string
	UploadFile(src, dst, host string, setHost bool) error
	Delete(path string) error
	MakeDirectory(path string) error
	Reference() types.ManagedObjectReference
}

type DatastoreDriver struct {
	ds     *object.Datastore
	driver *VCenterDriver
}

func (d *VCenterDriver) NewDatastore(ref *types.ManagedObjectReference) Datastore {
	return &DatastoreDriver{
		ds:     object.NewDatastore(d.client.Client, *ref),
		driver: d,
	}
}

// If name is an empty string, then resolve host's one
func (d *VCenterDriver) FindDatastore(name string, host string) (Datastore, error) {
	if name == "" {
		h, err := d.FindHost(host)
		if err != nil {
			return nil, fmt.Errorf("Error finding host for to get datastore: %s", err)
		}

		i, err := h.Info("datastore")
		if err != nil {
			return nil, fmt.Errorf("Error getting datastore info from host: %s", err)
		}

		if len(i.Datastore) > 1 {
			return nil, fmt.Errorf("Host has multiple datastores. Specify it explicitly")
		}

		ds := d.NewDatastore(&i.Datastore[0])
		inf, err := ds.Info("name")
		if err != nil {
			return nil, fmt.Errorf("Error getting datastore name: %s", err)
		}
		name = inf.Name
	}

	ds, err := d.finder.Datastore(d.ctx, name)
	if err != nil {
		return nil, fmt.Errorf("Error finding datastore with name %s: %s", name, err)
	}

	return &DatastoreDriver{
		ds:     ds,
		driver: d,
	}, nil
}

func (d *VCenterDriver) GetDatastoreName(id string) (string, error) {
	obj := types.ManagedObjectReference{
		Type:  "Datastore",
		Value: id,
	}
	pc := property.DefaultCollector(d.vimClient)
	var me mo.ManagedEntity

	err := pc.RetrieveOne(d.ctx, obj, []string{"name"}, &me)
	if err != nil {
		return id, err
	}
	return me.Name, nil
}

func (ds *DatastoreDriver) Info(params ...string) (*mo.Datastore, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.Datastore
	err := ds.ds.Properties(ds.driver.ctx, ds.ds.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (ds *DatastoreDriver) DirExists(filepath string) bool {
	_, err := ds.ds.Stat(ds.driver.ctx, filepath)
	if _, ok := err.(object.DatastoreNoSuchDirectoryError); ok {
		return false
	}
	return true
}

func (ds *DatastoreDriver) FileExists(path string) bool {
	_, err := ds.ds.Stat(ds.driver.ctx, path)
	return err == nil
}

func (ds *DatastoreDriver) Name() string {
	return ds.ds.Name()
}

func (ds *DatastoreDriver) Reference() types.ManagedObjectReference {
	return ds.ds.Reference()
}

func (ds *DatastoreDriver) ResolvePath(path string) string {
	return ds.ds.Path(path)
}

// The file ID isn't available via the API, so we use DatastoreBrowser to search
func (d *VCenterDriver) GetDatastoreFilePath(datastoreID, dir, filename string) (string, error) {
	ref := types.ManagedObjectReference{Type: "Datastore", Value: datastoreID}
	ds := object.NewDatastore(d.vimClient, ref)

	b, err := ds.Browser(d.ctx)
	if err != nil {
		return filename, err
	}
	ext := path.Ext(filename)
	pat := strings.Replace(filename, ext, "*"+ext, 1)
	spec := types.HostDatastoreBrowserSearchSpec{
		MatchPattern: []string{pat},
	}

	task, err := b.SearchDatastore(d.ctx, dir, &spec)
	if err != nil {
		return filename, err
	}

	info, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return filename, err
	}

	res, ok := info.Result.(types.HostDatastoreBrowserSearchResults)
	if !ok {
		return filename, fmt.Errorf("search(%s) result type=%T", pat, info.Result)
	}

	if len(res.File) != 1 {
		return filename, fmt.Errorf("search(%s) result files=%d", pat, len(res.File))
	}
	return res.File[0].GetFileInfo().Path, nil
}

func (ds *DatastoreDriver) UploadFile(src, dst, host string, setHost bool) error {
	p := soap.DefaultUpload
	ctx := ds.driver.ctx

	if setHost && host != "" {
		h, err := ds.driver.FindHost(host)
		if err != nil {
			return err
		}
		ctx = ds.ds.HostContext(ctx, h.host)
	}

	return ds.ds.UploadFile(ctx, src, dst, &p)
}

func (ds *DatastoreDriver) Delete(path string) error {
	dc, err := ds.driver.finder.Datacenter(ds.driver.ctx, ds.ds.DatacenterPath)
	if err != nil {
		return err
	}
	fm := ds.ds.NewFileManager(dc, false)
	return fm.Delete(ds.driver.ctx, path)
}

func (ds *DatastoreDriver) MakeDirectory(path string) error {
	dc, err := ds.driver.finder.Datacenter(ds.driver.ctx, ds.ds.DatacenterPath)
	if err != nil {
		return err
	}
	fm := ds.ds.NewFileManager(dc, false)
	return fm.FileManager.MakeDirectory(ds.driver.ctx, path, dc, true)
}

// Cuts out the datastore prefix
// Example: "[datastore1] file.ext" --> "file.ext"
func RemoveDatastorePrefix(path string) string {
	res := object.DatastorePath{}
	if hadPrefix := res.FromString(path); hadPrefix {
		return res.Path
	} else {
		return path
	}
}

type DatastoreIsoPath struct {
	path string
}

func (d *DatastoreIsoPath) Validate() bool {
	// Matches:
	// [datastore] /dir/subdir/file
	// [datastore] dir/subdir/file
	// [] /dir/subdir/file
	// [data-store] /dir/subdir/file
	// dir/subdir/file or dir/subdir/file
	matched, _ := regexp.MatchString(`^\s*(\[[^\[\]\/]*\])?\s*[^\[\]]+\s*$`, d.path)
	return matched
}

func (d *DatastoreIsoPath) GetFilePath() string {
	filePath := d.path
	parts := strings.Split(d.path, "]")
	if len(parts) > 1 {
		// removes datastore name from path
		filePath = parts[1]
		filePath = strings.TrimSpace(filePath)
	}
	return filePath
}
