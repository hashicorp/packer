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

func (b *Builder) ConfigInterface() interface{} {
	return &b.config
}

func (*Builder) Prepare() {
}

func (b *Builder) Build() {
}
