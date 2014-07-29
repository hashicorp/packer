package common

import (
        "errors" 
        "io/ioutil"
        "log"
        "os"
        "path/filepath"
        "strings"
        "github.com/walle/targz"
)

// The VagrantBox interface describes what a VagrantBox can do
type VagrantBox interface {
        Expand(vm_ext string) (string, error)
        Clean()
}

type DefaultVagrantBox struct {
        sourcePath               string
        vmPath                   string

        tempDir                  VagrantBoxTempDir
        expander                 VagrantBoxExpander
}

type VagrantBoxTempDir interface {
        Create() error
        Path() string
        ReadDir(targetPath string) ([]os.FileInfo, error)
        FindFileWithSuffix(suffix string) os.FileInfo
}

type DefaultVagrantBoxTempDir struct {
        path                     string
}
func (vbtd *DefaultVagrantBoxTempDir) Create() (error) {
        path, err := ioutil.TempDir("", "packer")
        vbtd.path = path
        return err
}
func (vbtd *DefaultVagrantBoxTempDir) Path() string {
        return vbtd.path
}
func (vbtd *DefaultVagrantBoxTempDir) ReadDir(targetPath string) ([]os.FileInfo, error) {
        return ioutil.ReadDir(targetPath) 
}
func (vbtd *DefaultVagrantBoxTempDir) FindFileWithSuffix(suffix string) os.FileInfo {
    files, _ := vbtd.ReadDir(vbtd.path)
    for _, fi := range files {
        if strings.HasSuffix(fi.Name(), suffix) {
            return fi
        }
    }
    return nil
}

type VagrantBoxExpander interface {
        Expand(sourcePath string, targetPath string) (error)
}

type DefaultVagrantBoxExpander struct {} 
func (vbe *DefaultVagrantBoxExpander) Expand(sourcePath string, targetPath string) (error) {
        return targz.Extract(sourcePath, targetPath)
}

// NewVagrantBox returns a new VagrantBox for the given
// configuration.
func NewVagrantBox(path string) VagrantBox {
        return &DefaultVagrantBox { 
                sourcePath: path,
                tempDir: &DefaultVagrantBoxTempDir{},
                expander: &DefaultVagrantBoxExpander{},
        }
}

// Expand takes the Vagrant-compatible box and uncompresses it into a
// directory. This function does not perform checks to verify that box is
// actually a proper box. This is an expected precondition.
func (vb *DefaultVagrantBox) Expand(vm_ext string) (string, error) {

        if strings.HasSuffix(vb.sourcePath, ".box") {

                log.Printf("Unboxing from: %s", vb.sourcePath)
                if err := vb.tempDir.Create(); err != nil {
                        return vb.sourcePath, err
                }

                if err := vb.expander.Expand(vb.sourcePath, vb.tempDir.Path()); err != nil {
                        return vb.sourcePath, err
                }
                log.Printf("Unboxed to: %s", vb.tempDir.Path())

                file := vb.tempDir.FindFileWithSuffix(vm_ext)

                if file == nil {
                        message := vm_ext + " not found in " + vb.tempDir.Path()
                        return vb.tempDir.Path(), errors.New(message)
                }

                vb.vmPath = filepath.Join(vb.tempDir.Path(), file.Name())
                return vb.vmPath, nil
        }

        return vb.sourcePath, nil
}

func (vb *DefaultVagrantBox) Clean() {
        if vb.tempDir.Path() != "" {
            os.RemoveAll(vb.tempDir.Path())
        }
}
