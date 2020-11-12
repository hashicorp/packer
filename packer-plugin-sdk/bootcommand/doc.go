/*
Package bootcommand generates and sends boot commands to the remote instance.

This package is relevant to people who want to create new builders, particularly
builders with the capacity to build a VM from an iso.

You can choose between three different drivers to send the command: a vnc
driver, a usb driver, and a PX-XT keyboard driver. The driver you choose will
depend on what kind of keyboard codes your hypervisor expects, and how you want
to implement the connection.
*/

package bootcommand
