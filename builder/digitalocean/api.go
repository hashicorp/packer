// All of the methods used to communicate with the digital_ocean API
// are here. Their API is on a path to V2, so just plain JSON is used
// in place of a proper client library for now.

package digitalocean

type Region struct {
	Slug string `json:"slug"`
	Name string `json:"name"`

	// v1 only
	Id uint `json:"id,omitempty"`

	// v2 only
	Sizes     []string `json:"sizes,omitempty"`
	Available bool     `json:"available,omitempty"`
	Features  []string `json:"features,omitempty"`
}

type RegionsResp struct {
	Regions []Region
}

type Size struct {
	Slug string `json:"slug"`

	// v1 only
	Id   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	// v2 only
	Memory       uint    `json:"memory,omitempty"`
	VCPUS        uint    `json:"vcpus,omitempty"`
	Disk         uint    `json:"disk,omitempty"`
	Transfer     float64 `json:"transfer,omitempty"`
	PriceMonthly float64 `json:"price_monthly,omitempty"`
	PriceHourly  float64 `json:"price_hourly,omitempty"`
}

type SizesResp struct {
	Sizes []Size
}

type Image struct {
	Id           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Distribution string `json:"distribution"`

	// v2 only
	Public    bool     `json:"public,omitempty"`
	ActionIds []string `json:"action_ids,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
}

type ImagesResp struct {
	Images []Image
}

type DigitalOceanClient interface {
	CreateKey(string, string) (uint, error)
	DestroyKey(uint) error
	CreateDroplet(string, string, string, string, uint, bool) (uint, error)
	DestroyDroplet(uint) error
	PowerOffDroplet(uint) error
	ShutdownDroplet(uint) error
	CreateSnapshot(uint, string) error
	Images() ([]Image, error)
	DestroyImage(uint) error
	DropletStatus(uint) (string, string, error)
	Image(string) (Image, error)
	Regions() ([]Region, error)
	Region(string) (Region, error)
	Sizes() ([]Size, error)
	Size(string) (Size, error)
}
