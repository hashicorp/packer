/*
Package adapter helps command line tools connect to the guest via a Packer
communicator. A typical use is for custom provisioners that wrap command line
tools. For example, the Ansible provisioner and the Inspec provisioner both
use this package to proxy communicator calls.
*/

package adapter
