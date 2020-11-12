/*
Package shell-local is designed to make it easier to shell out locally on the
machine running Packer. The top level tools in this package are probably not
relevant to plugin maintainers, as they are implementation details shared
between the HashiCorp-maintained shell-local provisioner and shell-local
post-processor.

The localexec sub-package can be used in any plugins that need local shell
access, whether that is in a driver for a hypervisor, or a command to a third
party cli tool. Please make sure that any third party tool dependencies are
noted in your plugin's documentation.
*/

package shell_local
