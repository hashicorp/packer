
communicator "ssh" "vagrant" {
  ssh_password                 = "s3cr4t"
  ssh_username                 = "vagrant"
  ssh_agent_auth               = false
  ssh_bastion_agent_auth       = true
  ssh_bastion_host             = ""
  ssh_bastion_password         = ""
  ssh_bastion_port             = 0
  ssh_bastion_private_key_file = ""
  ssh_bastion_username         = ""
  ssh_clear_authorized_keys    = true
  ssh_disable_agent_forwarding = true
  ssh_file_transfer_method     = "scp"
  ssh_handshake_attempts       = 32
  ssh_host                     = "sssssh.hashicorp.io"
  ssh_port                     = 42
  ssh_keep_alive_interval      = "10s"
  ssh_private_key_file         = "file.pem"
  ssh_proxy_host               = "ninja-potatoes.com"
  ssh_proxy_password           = "pickle-rick"
  ssh_proxy_port               = "42"
  ssh_proxy_username           = "dark-father"
  ssh_pty                      = false
  ssh_read_write_timeout       = "5m"
  ssh_timeout                  = "5m"
}
