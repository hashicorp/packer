---
description: |
    As Packer allows users to develop a custom builder as a plugin, NAVER CLOUD PLATFORM provides its own Packer builder for your convenience.
You can use NAVER CLOUD PLATFORM's Packer builder to easily create your server images.
layout: docs
page_title: 'Naver Cloud Platform - Builders'
sidebar_current: 'docs-builders-ncloud'
---

# NAVER CLOUD PLATFORM Builder

As Packer allows users to develop a custom builder as a plugin, NAVER CLOUD PLATFORM provides its own Packer builder for your convenience.
You can use NAVER CLOUD PLATFORM's Packer builder to easily create your server images.

#### Sample code of template.json

```
{
  "variables": {
    "ncloud_access_key": "FRxhOQRNjKVMqIz3sRLY",
    "ncloud_secret_key": "xd6kTO5iNcLookBx0D8TDKmpLj2ikxqEhc06MQD2"
  },
  "builders": [
    {
      "type": "ncloud",
      "access_key": "{{user `ncloud_access_key`}}",
      "secret_key": "{{user `ncloud_secret_key`}}",

      "server_image_product_code": "SPSW0WINNT000016",
      "server_product_code": "SPSVRSSD00000011",
      "member_server_image_no": "4223",
      "server_image_name": "packer-test {{timestamp}}",
      "server_description": "server description",
      "user_data": "CreateObject(\"WScript.Shell\").run(\"cmd.exe /c powershell Set-ExecutionPolicy RemoteSigned & winrm quickconfig -q & sc config WinRM start= auto & winrm set winrm/config/service/auth @{Basic=\"\"true\"\"} & winrm set winrm/config/service @{AllowUnencrypted=\"\"true\"\"} & winrm get winrm/config/service\")",
      "region": "US-West"
    }
  ]
}
```

#### Description

* type(required): "ncloud"
* ncloud_access_key (required): User's access key. Go to [[Account Management > Authentication Key]](https://www.ncloud.com/mypage/manage/authkey) to create and view your authentication key.
* ncloud_secret_key (required): User's secret key paired with the access key. Go to [[Account Management > Authentication Key]](https://www.ncloud.com/mypage/manage/authkey) to create and view your authentication key.
* server_image_product_code: Product code of an image to create. (member_server_image_no is required if not specified)
* server_product_code (required): Product (spec) code to create.
* member_server_image_no: Previous image code. If there is an image previously created, it can be used to create a new image. (server_image_product_code is required if not specified)
* server_image_name (option): Name of an image to create.
* server_image_description (option): Description of an image to create.
* block_storage_size (option): You can add block storage ranging from 10 GB to 2000 GB, in increments of 10 GB.
* access_control_group_configuration_no: This is used to allow winrm access when you create a Windows server. An ACG that specifies an access source ("0.0.0.0/0") and allowed port (5985) must be created in advance.
* user_data (option): Init script to run when an instance is created.
  * For Linux servers, Python, Perl, and Shell scripts can be used. The path of the script to run should be included at the beginning of the script, like #!/usr/bin/env python, #!/bin/perl, or #!/bin/bash.
  * For Windows servers, only Visual Basic scripts can be used.
  * All scripts must be written in English.
* region (option): Name of the region where you want to create an image. (default: Korea)
  * values: Korea / US-West / HongKong / Singapore / Japan / Germany

### Requirements for creating Windows images

You should include the following code in the packer configuration file for provision when creating a Windows server.

```
  "builders": [
    {
      "type": "ncloud",
      ...
      "user_data":
        "CreateObject(\"WScript.Shell\").run(\"cmd.exe /c powershell Set-ExecutionPolicy RemoteSigned & winrm set winrm/config/service/auth @{Basic=\"\"true\"\"} & winrm set winrm/config/service @{AllowUnencrypted=\"\"true\"\"} & winrm quickconfig -q & sc config WinRM start= auto & winrm get winrm/config/service\")",
      "communicator": "winrm",
      "winrm_username": "Administrator"
    }
  ],
  "provisioners": [
    {
      "type": "powershell",
      "inline": [
        "$Env:SystemRoot\\System32\\Sysprep\\Sysprep.exe /oobe /generalize /shutdown /quiet \"/unattend:C:\\Program Files (x86)\\NBP\\nserver64.xml\" "
      ]
    }
  ]
```

### Note

* You can only create as many public IP addresses as the number of server instances you own. Before running Packer, please make sure that the number of public IP addresses previously created is not larger than the number of server instances (including those to be used to create server images).
* When you forcibly terminate the packer process or close the terminal (command) window where the process is running, the resources may not be cleaned up as the packer process no longer runs. In this case, you should manually clean up the resources associated with the process.
