package yandeximport

import "fmt"

const sourceType_IMAGE = "image"
const sourceType_OBJECT = "object"

type cloudImageSource interface {
	GetSourceID() string
	GetSourceType() string
	Description() string
}

type imageSource struct {
	imageID string
}

func (i *imageSource) GetSourceID() string {
	return i.imageID
}

func (i *imageSource) GetSourceType() string {
	return sourceType_IMAGE
}

func (i *imageSource) Description() string {
	return fmt.Sprintf("%s source, id: %s", i.GetSourceType(), i.imageID)
}

type objectSource struct {
	url string
}

func (i *objectSource) GetSourceID() string {
	return i.url
}

func (i *objectSource) GetSourceType() string {
	return sourceType_OBJECT
}

func (i *objectSource) Description() string {
	return fmt.Sprintf("%s source, url: %s", i.GetSourceType(), i.url)
}
