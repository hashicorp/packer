package packer

// A Hook is used to hook into an arbitrarily named location in a build,
// allowing custom behavior to run at certain points along a build.
//
// Run is called when the hook is called, with the name of the hook and
// arbitrary data associated with it. To know what format the data is in,
// you must reference the documentation for the specific hook you're interested
// in. In addition to that, the Hook is given access to a UI so that it can
// output things to the user.
type Hook interface {
	Run(string, interface{}, Ui)
}
