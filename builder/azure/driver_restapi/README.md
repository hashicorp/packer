Azure builder
=============
* This is a builder to enable Microsoft Azure users to build custom images given an Azure image;
* The builder utilizes [Service Management REST API](http://msdn.microsoft.com/en-us/library/azure/ee460799.aspx) and ;
* You can execute the builder from both Windows and Linux dev-boxes (**clients**);
* You can build Linux and Windows Azure images (**targets**) with this builder:
  * to create **Windows target** you should use the builder together with **azure-custom-script-extension** provisioner;
	  * the provisioner utilizes [Storage Services REST API](http://msdn.microsoft.com/en-us/library/azure/dd179355.aspx);
	  * visit [the link](http://msdn.microsoft.com/en-us/library/dn781373.aspx) to understand how the provisioner works;
  * to create **Linux target** you can use any of the well known SSH base provisioners: **shell** and/or **file**;

# 1. Prerequisite

1. You must have an [Azure subscription](http://azure.microsoft.com) to begin using Azure;
2. Azure builder uses **OpenSSL** to manage certificates: 
	* If your are working on Linux you have **openssl** by default. To make sure it works try this command in a terminal: 
		openssl version 
	* If you are working on Windows you need to install OpenSSL. You have two options:
		1. install [OpenSSL for Windows](http://slproweb.com/products/Win32OpenSSL.html). 
			In this case you will be able to start Packer in any command prompt; 
		2. install [Git Bash](http://git-scm.com/downloads). 
			In this case you will be able to start Packer in Git Bash terminal only;
3. To start using the builder you will need to get [PublishSetting profile](http://go.microsoft.com/fwlink/?LinkId=254432);

# 2. Quick configuration examples

## a. Linux target
```javascript
{
	"variables": 
	{
		"psPath" 	: "your path to publishsettings file",
		"sn" 		: "your Azure subscription name",
		"an" 		: "your Azure storage account name",
		"cn" 		: "your Azure container name"
	},
	
	"builders":[{
		"type"						: "azure",
		"publish_settings_path" 	: "{{user `psPath`}}",
		"subscription_name"			: "{{user `sn`}}",
		"storage_account" 			: "{{user `an`}}",
		"storage_account_container" : "{{user `cn`}}",
		"os_type"					: "Linux",
		"os_image_label"			: "Ubuntu Server 14.04 LTS",
		"location"					: "West US",
		"instance_size"				: "Small",
		"user_image_label"			: "PackerMade_Ubuntu_Serv14"
	}],
	
	"provisioners":[{
		"type":	"shell",
		"execute_command":	"chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",
		"inline": [	"sudo apt-get update",
			"sudo apt-get install -y mc",
			"sudo apt-get install -y nodejs",
			"sudo apt-get install -y npm",
			"sudo npm install azure-cli -g"
		],
		"inline_shebang":	"/bin/sh -x"
	}] 
}			
```

## b. Windows target
```javascript
{
	"variables": 
	{
		"psPath" 	: "your path to publishsettings file",
		"sn" 		: "your Azure subscription name",
		"an" 		: "your Azure storage account name",
		"cn" 		: "your Azure container name"
	},
	
	"builders":[{
		"type"						: "azure",
		"publish_settings_path" 	: "{{user `psPath`}}",
		"subscription_name"			: "{{user `sn`}}",
		"storage_account" 			: "{{user `an`}}",
		"storage_account_container" : "{{user `cn`}}",
		"os_type"					: "Windows",
		"os_image_label"			: "Windows Server 2012 R2 Datacenter",
		"location"					: "West US",
		"instance_size"				: "Small",
		"user_image_label"			: "PackerMade_Windows2012R2DC"				
	}],
	
	"provisioners":[{
		"type":	"azure-custom-script-extension",
		"inline": [	
			"Write-Host 'Inline script!'",
			"Write-Host 'Installing Mozilla Firefox...'",
			"$filename = 'Firefox Setup 31.0.exe'",
			"$link = 'https://download.mozilla.org/?product=firefox-31.0-SSL&os=win&lang=en-US'",
			"$dstDir = 'c:/MyFileFolder'",
			"New-Item $dstDir -type directory -force | Out-Null",
			"$remotePath = Join-Path $dstDir $filename",
			"(New-Object System.Net.Webclient).downloadfile($link, $remotePath)",
			"Start-Process $remotePath -NoNewWindow -Wait -Argument '/S'",
			"Write-Host 'Inline script finished!'"
		]				
	}] 
}			
```
