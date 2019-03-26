package yandex

import "fmt"

type ArtifactMini struct {
	config      *Config
	imageID     string
	imageName   string
	imageFamily string
}

//revive:disable:var-naming
func (*ArtifactMini) BuilderId() string {
	return BuilderID
}

func (a *ArtifactMini) Id() string {
	return a.imageID
}

func (*ArtifactMini) Files() []string {
	return nil
}

//revive:enable:var-naming
func (a *ArtifactMini) String() string {
	return fmt.Sprintf("A disk image was created: %v (id: %v) (family: %v)", a.imageName, a.imageID, a.imageFamily)
}

func (a *ArtifactMini) State(name string) interface{} {
	switch name {
	case "ImageID":
		return a.imageID
	case "FolderID":
		return a.config.FolderID
	case "BuildZone":
		return a.config.Zone
	}
	return nil

}

func (*ArtifactMini) Destroy() error {
	// no destroy right now
	return nil
}
