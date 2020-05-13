package digitalocean

import (
	"testing"

	"github.com/digitalocean/godo"
)

func TestBuilder_GetImageType(t *testing.T) {
	imageTypeTests := []struct {
		in  string
		out godo.DropletCreateImage
	}{
		{"ubuntu-20-04-x64", godo.DropletCreateImage{Slug: "ubuntu-20-04-x64"}},
		{"123456", godo.DropletCreateImage{ID: 123456}},
	}

	for _, tt := range imageTypeTests {
		t.Run(tt.in, func(t *testing.T) {
			i := getImageType(tt.in)
			if i != tt.out {
				t.Errorf("got %q, want %q", godo.Stringify(i), godo.Stringify(tt.out))
			}
		})
	}
}
