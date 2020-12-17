/*
Package packer contains all of the interfaces for key Packer objects.

This module will need to be imported by all but the very simplest plugins. It
represents the foundation of the API that the Core and Plugins use to
communicate with each other.

Changes to any of the interfaces in this package likely represent a
backwards-incompatibility and should therefore only be made rarely and when
absolutely necessary.

Plugins will need to implement either the Builder, Provisioner,
or Post-Processor interfaces, and will likely create an Artifact. The
Communicator must be implemented in the Builder and then passed into the
Provisioners so they can use it communicate with the instance without needing
to know the connection details.

The UI is created by the Packer core for use by the plugins, and is how the
plugins stream information back to the terminal.
*/
package packer
