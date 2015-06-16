packer-hyperv
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is a Hyperv plugin for Packer.io to enable windows users to build custom images given an ISO. 

ISO's can be downloaded off technet or MSDN (if you have a subscription for the latter).
http://www.microsoft.com/en-us/evalcenter/evaluate-windows-server-2012-r2

The hyper-v plugin enables you to build a Windows Server Vagrant box for the hyper-v provider only. 

The bin folder has an example JSON to help specify the new hyperv configuration. 
      "type": "hyperv-iso",
			"guest_os_type":"WindowsServer2012R2Datacenter",
			"product_key" : "{{user `product_key`}}",
			"iso_url":"d:/Hyper-V/ISO/Windows_Server_2012_R2-EN-US-x64.ISO",

Additionally, as indicated above, if you obtain a windows license, you can specify the product key within your .json configuration and the plugin will register your copy of windows. 

Note: The plugin has to be run on a Windows workstation 8.1 or higher and must have hyper-v enabled. 
