package yandex

type Image struct {
	ID            string
	FolderID      string
	Labels        map[string]string
	Licenses      []string
	MinDiskSizeGb int
	Name          string
	Description   string
	Family        string
	SizeGb        int
}
