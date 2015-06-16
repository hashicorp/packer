package response

import (
	"io"
	"encoding/xml"
)

type ContainersList struct {
	XMLName   			xml.Name `xml:"EnumerationResults"`
	ServiceEndpoint	  	string `xml:"ServiceEndpoint,attr"`
	Prefix  			string
	Marker  			string
	MaxResults  		int

	Containers 			[]Container `xml:"Containers>Container"`
	NextMarker  		string
}

type Container struct {
	Name  			string
	Url  			string
	Properties		[]Property
	Metadata		string
}

type Property struct {
	LastModified  	string	`xml:"Last-Modified"`
	Etag  			string
}


func ParseContainersList(body io.ReadCloser) (*ContainersList, error ) {
	data, err := toModel(body, &ContainersList{})

	if err != nil {
		return nil, err
	}

	m := data.(*ContainersList)

	return m, nil
}
