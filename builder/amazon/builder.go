package amazon

type config struct {
	AccessKey string
	Region    string
	SecretKey string
	SourceAmi string
}

type Builder struct {
	config config
}

func (*Builder) Prepare() {
}

func (*Builder) Build() {
}

func (*Builder) Destroy() {
}
