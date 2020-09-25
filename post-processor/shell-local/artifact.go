package shell_local

type Artifact struct {
	builderId string
	stringVal string
	destroy   func() error
	files     []string
	id        string
	state     func(name string) interface{}
}

func (a *Artifact) BuilderId() string {
	return a.builderId
}

func (a *Artifact) Files() []string {
	return a.files
}

func (a *Artifact) Id() string {
	return a.id
}

func (a *Artifact) String() string {
	return a.stringVal
}

func (a *Artifact) State(name string) interface{} {
	return a.state(name)
}

func (a *Artifact) Destroy() error {
	return a.destroy()
}
