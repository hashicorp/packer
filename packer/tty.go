package packer

type TTY interface {
	ReadString() (string, error)
	Close() error
}
