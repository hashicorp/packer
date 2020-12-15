/*
Package communicator provides common steps for connecting to an instance
using the Packer communicator. These steps can be implemented by builders.
Normally, a builder will want to implement StepConnect, which is smart enough
to then determine which kind of communicator, and therefore which kind of
substep, it should implement.

Various helper functions are also supplied.
*/
package communicator
