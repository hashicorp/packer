## Which SSH Options to use:

This chart breaks down what Packer does if you set any of the below SSH options:

| ssh_password | ssh_private_key_file | ssh_keypair_name | temporary_key_pair_name | Packer will... |
| --- | --- | --- | --- | --- |
| X | - | - | - | ssh authenticating with username and given password |
| - | X | - | - | ssh authenticating with private key file |
| - | X | X | - | ssh authenticating with given private key file and "attaching" the keypair to the instance |
| - | - | - | X | Create a temporary ssh keypair with a particular name, clean it up |
| - | - | - | - | Create a temporary ssh keypair with a default name, clean it up |
