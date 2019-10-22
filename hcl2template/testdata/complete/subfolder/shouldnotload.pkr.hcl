
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
        communicator = comm.ssh.vagrant

        shell {
            inline = [
                "echo '{{user `my_secret`}}' :D"
            ]
        }

        shell {
            script = [
                "script-1.sh",
                "script-2.sh",
            ]
            override "vmware-iso" {
                execute_command = "echo 'password' | sudo -S bash {{.Path}}"
            }
        }

        upload "log.go" "/tmp" {
            timeout = "5s"
        }

    }

    post_provision {
        amazon-import {
            only = ["src.virtualbox-iso.ubuntu-1204"]
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
        communicator = comm.ssh.vagrant

        shell {
            inline = [
                "echo HOLY GUACAMOLE !"
            ]
        }
    }
}