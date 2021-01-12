/*
The commonsteps package contains the multistep runner that comprises the main
architectural convention of Packer builder plugins. It enables builders
to respect global Packer flags like "on-error" and "debug". It also contains
a selection of convenience "multistep" steps that perform globally relevant
tasks that many or most builders will want to implement -- for example,
launching Packer's internal HTTP server for serving files to the instance.

It also provides step_provision, which contains the hooks necessary for allowing
provisioners to run inside your builder.

While it is possible to create a simple builder without using the multistep
runner or step_provision, your builder will lack core Packer functionality.
*/
package commonsteps
