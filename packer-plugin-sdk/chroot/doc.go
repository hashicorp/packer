/*
Package chroot provides convenience tooling specific to chroot builders.

Chroot builders work by creating a new volume from an existing source image and
attaching it into an already-running instance. Once attached, a chroot is used
to provision the system within that volume. After provisioning, the volume is
detached, snapshotted, and a cloud-specific image is made.

Using this process, minutes can be shaved off image build processes because a
new instance doesn't need to be launched in the cloud before provisioning can
take place.

There are some restrictions, however. The host instance where the volume is
attached to must be a similar system (generally the same OS version, kernel
versions, etc.) as the image being built. Additionally, this process is much
more expensive because the instance used to perform the build must be kept
running persistently in order to build images, whereas the other non-chroot
cloud image builders start instances on-demand to build images as needed.

The HashiCorp-maintained Amazon and Azure builder plugins have chroot builders
which use this option and can serve as an example for how the chroot steps and
communicator are used.
*/
package chroot
