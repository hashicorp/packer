/*
Package guestexec provides a shim for running common operating system commands
on the guest/remote instance that is being provisioned. It helps provisioners
which need to perform operating-system specific calls do so in a way that is
simple and repeatable.

Note that to successfully use this package your provisioner must have knowledge
of the guest type, which is not information that builders generally collect --
your provisioner will have to require guest information in its config.
*/

package guestexec
