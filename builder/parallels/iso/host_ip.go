package iso

// Interface to help find the host IP that is available from within
// the Parallels virtual machines.
type HostIPFinder interface {
	HostIP() (string, error)
}
