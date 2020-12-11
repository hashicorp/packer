/*
Package adapter helps command line tools connect to the guest via a Packer
communicator.

A typical use is for custom provisioners that wrap command line
tools. For example, the Ansible provisioner and the Inspec provisioner both
use this package to proxy communicator calls.

You may want to use this adapter if you are writing a provisioner that wraps a
tool which under normal usage would be run locally and form a connection to the
remote instance itself.
*/

package adapter
