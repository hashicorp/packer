package command

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/hashicorp/harmony-go/archive"
	"github.com/mitchellh/packer/packer"
)

// archiveTemplateEntry is the name the template always takes within the slug.
const archiveTemplateEntry = ".packer-template.json"

type PushCommand struct {
	Meta

	// For tests:
	uploadFn func(io.Reader, *uploadOpts) (<-chan struct{}, <-chan error, error)
}

func (c *PushCommand) Run(args []string) int {
	var token string

	f := flag.NewFlagSet("push", flag.ContinueOnError)
	f.Usage = func() { c.Ui.Error(c.Help()) }
	f.StringVar(&token, "token", "", "token")
	if err := f.Parse(args); err != nil {
		return 1
	}

	args = f.Args()
	if len(args) != 1 {
		f.Usage()
		return 1
	}

	// Read the template
	tpl, err := packer.ParseTemplateFile(args[0], nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	// Validate some things
	if tpl.Push.Name == "" {
		c.Ui.Error(fmt.Sprintf(
			"The 'push' section must be specified in the template with\n" +
				"at least the 'name' option set."))
		return 1
	}

	// Build the archiving options
	var opts archive.ArchiveOpts
	opts.Include = tpl.Push.Include
	opts.Exclude = tpl.Push.Exclude
	opts.VCS = tpl.Push.VCS
	opts.Extra = map[string]string{
		archiveTemplateEntry: args[0],
	}

	// Determine the path we're archiving
	path := tpl.Push.BaseDir
	if path == "" {
		path, err = filepath.Abs(args[0])
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error determining path to archive: %s", err))
			return 1
		}
		path = filepath.Dir(path)
	}

	// Build the upload options
	var uploadOpts uploadOpts
	uploadOpts.Slug = tpl.Push.Name
	uploadOpts.Token = token

	// Start the archiving process
	r, archiveErrCh, err := archive.Archive(path, &opts)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error archiving: %s", err))
		return 1
	}
	defer r.Close()

	// Start the upload process
	doneCh, uploadErrCh, err := c.upload(r, &uploadOpts)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error starting upload: %s", err))
		return 1
	}

	err = nil
	select {
	case err = <-archiveErrCh:
		err = fmt.Errorf("Error archiving: %s", err)
	case err = <-uploadErrCh:
		err = fmt.Errorf("Error uploading: %s", err)
	case <-doneCh:
	}

	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}

func (*PushCommand) Help() string {
	helpText := `
Usage: packer push [options] TEMPLATE

  Push the template and the files it needs to a Packer build service.
  This will not initiate any builds, it will only update the templates
  used for builds.

Options:

  -token=<token>      Access token to use to upload. If blank, the
                      TODO environmental variable will be used.
`

	return strings.TrimSpace(helpText)
}

func (*PushCommand) Synopsis() string {
	return "push template files to a Packer build service"
}

func (c *PushCommand) upload(
	r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
	if c.uploadFn != nil {
		return c.uploadFn(r, opts)
	}

	return nil, nil, nil
}

type uploadOpts struct {
	URL   string
	Slug  string
	Token string
}
