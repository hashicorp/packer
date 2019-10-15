
// starts resources to provision them.
build {
    from "src.amazon-ebs.ubuntu-1604" {
        ami_name = "that-ubuntu-1.0"
    }

    from "src.virtualbox-iso.ubuntu-1204" {
        // build name is defaulted from the label "src.virtualbox-iso.ubuntu-1204"
        outout_dir = "path/"
    }

    provision {
        communicator = "comm.ssh.vagrant"

        shell {
            inline = [
                "echo '{{user `my_secret`}}' :D"
            ]
        }

        shell {
            valid_exit_codes = [
                0,
                42,
            ]
            scripts = [
                "script-1.sh",
                "script-2.sh",
            ]
            // override "vmware-iso" { // TODO(azr): handle common fields
            //     execute_command = "echo 'password' | sudo -S bash {{.Path}}"
            // }
        }

        file {
            source = "app.tar.gz"
            destination = "/tmp/app.tar.gz"
            // timeout = "5s" // TODO(azr): handle common fields
        }

    }

    post_provision {
        amazon-import {
            // only = ["src.virtualbox-iso.ubuntu-1204"] // TODO(azr): handle common fields
            ami_name = "that-ubuntu-1.0"
        }
    }
}

build {
    // build an ami using the ami from the previous build block.
    from "src.amazon.that-ubuntu-1.0" {
        ami_name = "fooooobaaaar"
    }

    provision {

        shell {
            inline = [
                "echo HOLY GUACAMOLE !"
            ]
        }
    }
}