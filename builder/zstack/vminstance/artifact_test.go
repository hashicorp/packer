package vminstance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"
	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}

func TestArtifactId(t *testing.T) {
	exportPaths := []string{"/zstack_export/image1", "/zstack_export/image2"}

	image1 := &zstacktype.Image{
		Name:          "foo1-Data",
		Uuid:          "uuid1",
		BackupStorage: "bs",
		Type:          "zstack",
		Platform:      "Linux",
	}
	image2 := &zstacktype.Image{
		Name:          "foo2-Root",
		Uuid:          "uuid2",
		BackupStorage: "bs",
		Type:          "zstack",
		Platform:      "Linux",
	}

	a := &Artifact{
		builderIdValue: BuilderId,
		driver:         *new(Driver),
		config:         *new(Config),
		images:         []*zstacktype.Image{image1, image2},
		exportPath:     exportPaths,
	}
	if len(a.Files()) != 2 {
		t.Fatalf("should export 2 paths")
	}

	if a.Id() != "uuid2" {
		t.Fatalf(fmt.Sprintf("Id should be Root Image Id, which is [%s]", "uuid2"))
	}
}
