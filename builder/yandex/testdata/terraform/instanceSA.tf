// This terraform code allows you to quickly create an instance with the assigned service account and the necessary rights.
// To start, it is required to specify two parameters: oauth token and folder id.
//
// $ terraform apply -var yc_token=<your_token_value> -var folder_id=<your_folder_id>
//
// After testing and completing work, just run the command to delete all provisioned cloud resources:
//
// $ terraform destroy -var yc_token=<your_token_value> -var folder_id=<your_folder_id>
//

// Variables section
variable "yc_token" {
  description = "Yandex.Cloud OAuth token"
}

variable "folder_id" {
  description = "Folder ID to provision all cloud resources"
}

variable "username" {
  default = "ubuntu"
}

variable "path_to_ssh_public_key" {
  default = "~/.ssh/id_rsa.pub"
}

// Provider section
provider "yandex" {
  folder_id = var.folder_id
  token     = var.yc_token
  zone      = "ru-central1-a"
}

// Main section
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

resource "yandex_vpc_network" "this" {
}

resource "yandex_vpc_subnet" "this" {
  network_id     = yandex_vpc_network.this.id
  v4_cidr_blocks = ["192.168.86.0/24"]
}

resource "yandex_compute_instance" "this" {
  service_account_id = yandex_iam_service_account.this.id
  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
    }
  }
  network_interface {
    subnet_id = yandex_vpc_subnet.this.id
    nat       = true
  }
  resources {
    cores         = 1
    memory        = 1
    core_fraction = 20
  }
  metadata = {
    ssh-keys = "${var.username}:${file(var.path_to_ssh_public_key)}"
  }
}

resource "yandex_iam_service_account" "this" {
  name = "test-sa-for-instance"
}

resource "yandex_resourcemanager_folder_iam_member" "this" {
  folder_id = var.folder_id
  member    = "serviceAccount:${yandex_iam_service_account.this.id}"
  role      = "editor"
}

// Output section
output "result" {
  value = "\nuse ssh with login `ubuntu` to connect instance like:\n\n$ ssh -i ${var.path_to_ssh_public_key} -l ${var.username} ${yandex_compute_instance.this.network_interface[0].nat_ip_address}"
}
