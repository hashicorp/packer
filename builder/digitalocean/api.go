// All of the methods used to communicate with the digital_ocean API
// are here. Their API is on a path to V2, so just plain JSON is used
// in place of a proper client library for now.

package digitalocean

type Region struct {
	Id        uint     `json:"id,omitempty"`        //only in v1 api
	Slug      string   `json:"slug"`                //presen in both api
	Name      string   `json:"name"`                //presen in both api
	Sizes     []string `json:"sizes,omitempty"`     //only in v2 api
	Available bool     `json:"available,omitempty"` //only in v2 api
	Features  []string `json:"features,omitempty"`  //only in v2 api
}

type RegionsResp struct {
	Regions []Region
}

type Size struct {
	Id           uint     `json:"id,omitempty"`            //only in v1 api
	Name         string   `json:"name,omitempty"`          //only in v1 api
	Slug         string   `json:"slug"`                    //presen in both api
	Memory       uint     `json:"memory,omitempty"`        //only in v2 api
	VCPUS        uint     `json:"vcpus,omitempty"`         //only in v2 api
	Disk         uint     `json:"disk,omitempty"`          //only in v2 api
	Transfer     float64  `json:"transfer,omitempty"`      //only in v2 api
	PriceMonthly float64  `json:"price_monthly,omitempty"` //only in v2 api
	PriceHourly  float64  `json:"price_hourly,omitempty"`  //only in v2 api
	Regions      []string `json:"regions,omitempty"`       //only in v2 api
}

type SizesResp struct {
	Sizes []Size
}

type Image struct {
	Id           uint     `json:"id"`                   //presen in both api
	Name         string   `json:"name"`                 //presen in both api
	Slug         string   `json:"slug"`                 //presen in both api
	Distribution string   `json:"distribution"`         //presen in both api
	Public       bool     `json:"public,omitempty"`     //only in v2 api
	Regions      []string `json:"regions,omitempty"`    //only in v2 api
	ActionIds    []string `json:"action_ids,omitempty"` //only in v2 api
	CreatedAt    string   `json:"created_at,omitempty"` //only in v2 api
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
