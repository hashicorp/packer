package common

import (
	"testing"
)

func TestArtifact_String(t *testing.T) {
	a := &Artifact{
		Resources: []string{
			"/subscriptions/4674464f-6024-43ae-903c-f6eed761be04/resourceGroups/rg/providers/Microsoft.Compute/disks/PackerTemp-osdisk-1586461959",
			"/subscriptions/4674464f-6024-43ae-903c-f6eed761be04/resourceGroups/images/providers/Microsoft.Compute/galleries/testgallery/images/myUbuntu/versions/1.0.10",
		},
	}
	want := `Azure resources created:
/subscriptions/4674464f-6024-43ae-903c-f6eed761be04/resourcegroups/images/providers/microsoft.compute/galleries/testgallery/images/myubuntu/versions/1.0.10
/subscriptions/4674464f-6024-43ae-903c-f6eed761be04/resourcegroups/rg/providers/microsoft.compute/disks/packertemp-osdisk-1586461959
`
	if got := a.String(); got != want {
		t.Errorf("Artifact.String() = %v, want %v", got, want)
	}
}
