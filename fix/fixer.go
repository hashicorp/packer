package fix

// A Fixer is something that can perform a fix operation on a template.
type Fixer interface {
	// Fix takes a raw map structure input, potentially transforms it
	// in some way, and returns the new, transformed structure. The
	// Fix method is allowed to mutate the input.
	Fix(input map[string]interface{}) (map[string]interface{}, error)

	// Synopsis returns a string description of what the fixer actually
	// does.
	Synopsis() string
}

// Fixers is the map of all available fixers, by name.
var Fixers map[string]Fixer

// FixerOrder is the default order the fixers should be run.
var FixerOrder []string

func init() {
	Fixers = map[string]Fixer{
		"iso-md5":                new(FixerISOMD5),
		"createtime":             new(FixerCreateTime),
		"pp-vagrant-override":    new(FixerVagrantPPOverride),
		"virtualbox-gaattach":    new(FixerVirtualBoxGAAttach),
		"virtualbox-rename":      new(FixerVirtualBoxRename),
		"vmware-rename":          new(FixerVMwareRename),
		"parallels-headless":     new(FixerParallelsHeadless),
		"parallels-deprecations": new(FixerParallelsDeprecations),
		"sshkeypath":             new(FixerSSHKeyPath),
	}

	FixerOrder = []string{
		"iso-md5",
		"createtime",
		"virtualbox-gaattach",
		"pp-vagrant-override",
		"virtualbox-rename",
		"vmware-rename",
		"parallels-headless",
		"parallels-deprecations",
		"sshkeypath",
	}
}
