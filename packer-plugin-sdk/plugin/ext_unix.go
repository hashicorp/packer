// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package plugin

var (
	FileExtension = "_x" + APIVersion // OS-Specific plugin file extention
)
