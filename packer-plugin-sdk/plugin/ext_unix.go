// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package plugin

var (
	FileExtension = ".0_x" + APIVersion // OS-Specific plugin file extention
)
