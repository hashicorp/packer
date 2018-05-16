---
layout: intro
sidebar_current: intro-getting-started-build-image
page_title: Build an Image - Getting Started
description: |-
  With Packer installed, let's just dive right into it and build our first
  image. Our first image will be an Amazon EC2 AMI with Redis pre-installed.
  This is just an example. Packer can create images for many platforms with
  anything pre-installed.
---

# Build an Image

With Packer installed, let's just dive right into it and build our first image.
Our first image will be an [Amazon EC2 AMI](https://aws.amazon.com/ec2/) 
This is just an example. Packer can create images for [many platforms][platforms].

If you don't have an AWS account, [create one now](https://aws.amazon.com/free/).
For the example, we'll use a "t2.micro" instance to build our image, which
qualifies under the AWS [free-tier](https://aws.amazon.com/free/), meaning it
will be free. If you already have an AWS account, you may be charged some amount
of money, but it shouldn't be more than a few cents.

-> **Note:** If you're not using an account that qualifies under the AWS
free-tier, you may be charged to run these examples. The charge should only be a
few cents, but we're not responsible if it ends up being more.

Packer can build images for [many platforms][platforms] other than
AWS, but AWS requires no additional software installed on your computer and
their [free-tier](https://aws.amazon.com/free/) makes it free to use for most
people. This is why we chose to use AWS for the example. If you're uncomfortable
setting up an AWS account, feel free to follow along as the basic principles
apply to the other platforms as well.

## The Template

The configuration file used to define what image we want built and how is called
a *template* in Packer terminology. The format of a template is simple
[JSON](http://www.json.org/). JSON struck the best balance between
human-editable and machine-editable, allowing both hand-made templates as well
as machine generated templates to easily be made.

We'll start by creating the entire template, then we'll go over each section
briefly. Create a file `example.json` and fill it with the following contents:

```json
{
  "variables": {
    "aws_access_key": "",
    "aws_secret_key": ""
  },
  "builders": [{
    "type": "amazon-ebs",
    "access_key": "{{user `aws_access_key`}}",
    "secret_key": "{{user `aws_secret_key`}}",
    "region": "us-east-1",
    "source_ami_filter": {
      "filters": {
      "virtualization-type": "hvm",
      "name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
      "root-device-type": "ebs"
      },
      "owners": ["099720109477"],
      "most_recent": true
    },
    "instance_type": "t2.micro",
    "ssh_username": "ubuntu",
    "ami_name": "packer-example {{timestamp}}"
  }]
}
```

When building, you'll pass in `aws_access_key` and `aws_secret_key` as
[user variables](/docs/templates/user-variables.html), keeping your secret keys
out of the template. You can create security credentials on [this
page](https://console.aws.amazon.com/iam/home?#security_credential). An example
IAM policy document can be found in the [Amazon EC2 builder
docs](/docs/builders/amazon.html).

This is a basic template that is ready-to-go. It should be immediately
recognizable as a normal, basic JSON object. Within the object, the `builders`
section contains an array of JSON objects configuring a specific *builder*. A
builder is a component of Packer that is responsible for creating a machine and
turning that machine into an image.

In this case, we're only configuring a single builder of type `amazon-ebs`. This
is the Amazon EC2 AMI builder that ships with Packer. This builder builds an
EBS-backed AMI by launching a source AMI, provisioning on top of that, and
re-packaging it into a new AMI.

The additional keys within the object are configuration for this builder,
specifying things such as access keys, the source AMI to build from and more.
The exact set of configuration variables available for a builder are specific to
each builder and can be found within the [documentation](/docs/index.html).

Before we take this template and build an image from it, let's validate the
template by running `packer validate example.json`. This command checks the
syntax as well as the configuration values to verify they look valid. The output
should look similar to below, because the template should be valid. If there are
any errors, this command will tell you.

```text
$ packer validate example.json
Template validated successfully.
```

Next, let's build the image from this template.

An astute reader may notice that we said earlier we'd be building an image with
Redis pre-installed, and yet the template we made doesn't reference Redis
anywhere. In fact, this part of the documentation will only cover making a first
basic, non-provisioned image. The next section on provisioning will cover
installing Redis.

## Your First Image

With a properly validated template. It is time to build your first image. This
is done by calling `packer build` with the template file. The output should look
similar to below. Note that this process typically takes a few minutes.

-> **Note:** For the tutorial it is convenient to use the credentials in the
command line. However, it is potentially insecure. See our documentation for
other ways to [specify Amazon credentials](/docs/builders/amazon.html#specifying-amazon-credentials).

-> **Note:** When using packer on Windows, replace the single-quotes in the
command below with double-quotes.

```text
$ packer build \
    -var 'aws_access_key=YOUR ACCESS KEY' \
    -var 'aws_secret_key=YOUR SECRET KEY' \
    example.json
==> amazon-ebs: amazon-ebs output will be in this color.

==> amazon-ebs: Creating temporary keypair for this instance...
==> amazon-ebs: Creating temporary security group for this instance...
==> amazon-ebs: Authorizing SSH access on the temporary security group...
==> amazon-ebs: Launching a source AWS instance...
==> amazon-ebs: Waiting for instance to become ready...
==> amazon-ebs: Connecting to the instance via SSH...
==> amazon-ebs: Stopping the source instance...
==> amazon-ebs: Waiting for the instance to stop...
==> amazon-ebs: Creating the AMI: packer-example 1371856345
==> amazon-ebs: AMI: ami-19601070
==> amazon-ebs: Waiting for AMI to become ready...
==> amazon-ebs: Terminating the source AWS instance...
==> amazon-ebs: Deleting temporary security group...
==> amazon-ebs: Deleting temporary keypair...
==> amazon-ebs: Build finished.

==> Builds finished. The artifacts of successful builds are:
--> amazon-ebs: AMIs were created:

us-east-1: ami-19601070
```

At the end of running `packer build`, Packer outputs the *artifacts* that were
created as part of the build. Artifacts are the results of a build, and
typically represent an ID (such as in the case of an AMI) or a set of files
(such as for a VMware virtual machine). In this example, we only have a single
artifact: the AMI in us-east-1 that was created.

This AMI is ready to use. If you wanted you could go and launch this AMI right 
now and it would work great.

-> **Note:** Your AMI ID will surely be different than the one above. If you
try to launch the one in the example output above, you will get an error. If you
want to try to launch your AMI, get the ID from the Packer output.

-> **Note:** If you see a `VPCResourceNotSpecified` error, Packer might not be
able to determine the default VPC, which the `t2` instance types require. This
can happen if you created your AWS account before `2013-12-04`.  You can either
change the `instance_type` to `m3.medium`, or specify a VPC. Please see
http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html for more
information. If you specify a `vpc_id`, you will also need to set `subnet_id`.
Unless you modify your subnet's [IPv4 public addressing attribute](
http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/vpc-ip-addressing.html#subnet-public-ip),
you will also need to set `associate_public_ip_address` to `true`, or set up a
[VPN](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_VPN.html).

## Managing the Image

Packer only builds images. It does not attempt to manage them in any way. After
they're built, it is up to you to launch or destroy them as you see fit. If you
want to store and namespace images for quick reference, you can use [Atlas by
HashiCorp](https://atlas.hashicorp.com). We'll cover remotely building and
storing images at the end of this getting started guide.

After running the above example, your AWS account now has an AMI associated with
it. AMIs are stored in S3 by Amazon, so unless you want to be charged about
$0.01 per month, you'll probably want to remove it. Remove the AMI by first
deregistering it on the [AWS AMI management
page](https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Images). Next,
delete the associated snapshot on the [AWS snapshot management
page](https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Snapshots).

Congratulations! You've just built your first image with Packer. Although the
image was pretty useless in this case (nothing was changed about it), this page
should've given you a general idea of how Packer works, what templates are and
how to validate and build templates into machine images.

## Some more examples:

### Another Linux Example, with provisioners:
Create a file named `welcome.txt` and add the following:
```
WELCOME TO PACKER!
```

Create a file named `example.sh` and add the following:
```
#!/bin/bash
echo "hello
```

Set your access key and id as environment variables, so we don't need to pass 
them in through the command line:
```
export AWS_ACCESS_KEY_ID=MYACCESSKEYID
export AWS_SECRET_ACCESS_KEY=MYSECRETACCESSKEY
```

Now save the following text in a file named `firstrun.json`:

```
{
    "variables": {
        "aws_access_key": "{{env `AWS_ACCESS_KEY_ID`}}",
        "aws_secret_key": "{{env `AWS_SECRET_ACCESS_KEY`}}",
        "region":         "us-east-1"
    },
    "builders": [
        {
            "access_key": "{{user `aws_access_key`}}",
            "ami_name": "packer-linux-aws-demo-{{timestamp}}",
            "instance_type": "t2.micro",
            "region": "us-east-1",
            "secret_key": "{{user `aws_secret_key`}}",
            "source_ami_filter": {
              "filters": {
              "virtualization-type": "hvm",
              "name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
              "root-device-type": "ebs"
              },
              "owners": ["099720109477"],
              "most_recent": true
            },
            "ssh_username": "ubuntu",
            "type": "amazon-ebs"
        }
    ],
    "provisioners": [
        {
            "type": "file",
            "source": "./welcome.txt",
            "destination": "/home/ubuntu/"
        },
        {
            "type": "shell",
            "inline":[
                "ls -al /home/ubuntu",
                "cat /home/ubuntu/welcome.txt"
            ]
        },
        {
            "type": "shell",
            "script": "./example.sh"
        }
    ]
}
```

and to build, run `packer build firstrun.json`

Note that if you wanted to use a `source_ami` instead of a `source_ami_filter` 
it might look something like this: `"source_ami": "ami-fce3c696",`

Your output will look like this: 

```
amazon-ebs output will be in this color.

==> amazon-ebs: Prevalidating AMI Name: packer-linux-aws-demo-1507231105
    amazon-ebs: Found Image ID: ami-fce3c696
==> amazon-ebs: Creating temporary keypair: packer_59d68581-e3e6-eb35-4ae3-c98d55cfa04f
==> amazon-ebs: Creating temporary security group for this instance: packer_59d68584-cf8a-d0af-ad82-e058593945ea
==> amazon-ebs: Authorizing access to port 22 on the temporary security group...
==> amazon-ebs: Launching a source AWS instance...
==> amazon-ebs: Adding tags to source instance
    amazon-ebs: Adding tag: "Name": "Packer Builder"
    amazon-ebs: Instance ID: i-013e8fb2ced4d714c
==> amazon-ebs: Waiting for instance (i-013e8fb2ced4d714c) to become ready...
==> amazon-ebs: Waiting for SSH to become available...
==> amazon-ebs: Connected to SSH!
==> amazon-ebs: Uploading ./scripts/welcome.txt => /home/ubuntu/
==> amazon-ebs: Provisioning with shell script: /var/folders/8t/0yb5q0_x6mb2jldqq_vjn3lr0000gn/T/packer-shell661094204
    amazon-ebs: total 32
    amazon-ebs: drwxr-xr-x 4 ubuntu ubuntu 4096 Oct  5 19:19 .
    amazon-ebs: drwxr-xr-x 3 root   root   4096 Oct  5 19:19 ..
    amazon-ebs: -rw-r--r-- 1 ubuntu ubuntu  220 Apr  9  2014 .bash_logout
    amazon-ebs: -rw-r--r-- 1 ubuntu ubuntu 3637 Apr  9  2014 .bashrc
    amazon-ebs: drwx------ 2 ubuntu ubuntu 4096 Oct  5 19:19 .cache
    amazon-ebs: -rw-r--r-- 1 ubuntu ubuntu  675 Apr  9  2014 .profile
    amazon-ebs: drwx------ 2 ubuntu ubuntu 4096 Oct  5 19:19 .ssh
    amazon-ebs: -rw-r--r-- 1 ubuntu ubuntu   18 Oct  5 19:19 welcome.txt
    amazon-ebs: WELCOME TO PACKER!
==> amazon-ebs: Provisioning with shell script: ./example.sh
    amazon-ebs: hello
==> amazon-ebs: Stopping the source instance...
    amazon-ebs: Stopping instance, attempt 1
==> amazon-ebs: Waiting for the instance to stop...
==> amazon-ebs: Creating the AMI: packer-linux-aws-demo-1507231105
    amazon-ebs: AMI: ami-f76ea98d
==> amazon-ebs: Waiting for AMI to become ready...
```

### A windows example

Note that this uses a larger instance.  You will be charged for it. Also keep 
in mind that using windows AMIs incurs a fee that you don't get when you use 
linux AMIs.

You'll need to have a boostrapping file to enable ssh or winrm; here's a basic 
example of that file.

```
# set administrator password
net user Administrator SuperS3cr3t!
wmic useraccount where "name='Administrator'" set PasswordExpires=FALSE

# First, make sure WinRM doesn't run and can't be connected to
netsh advfirewall firewall add rule name="WinRM" protocol=TCP dir=in localport=5985 action=block
net stop winrm

# turn off PowerShell execution policy restrictions
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope LocalMachine

# configure WinRM
winrm quickconfig -q
winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="0"}'
winrm set winrm/config '@{MaxTimeoutms="7200000"}'
winrm set winrm/config/service '@{AllowUnencrypted="true"}'
winrm set winrm/config/service '@{MaxConcurrentOperationsPerUser="12000"}'
winrm set winrm/config/service/auth '@{Basic="true"}'
winrm set winrm/config/client/auth '@{Basic="true"}'

net stop winrm
set-service winrm -startupType automatic

# Finally, allow WinRM connections and start the service
netsh advfirewall firewall set rule name="WinRM" new action=allow
net start winrm
```


Save the above code in a file named `bootstrap_win.txt`.  

The example config below shows the two different ways of using the powershell 
provisioner: `inline` and `script`.  
The first example, `inline`, allows you to provide short snippets of code, and 
will create the script file for you.  The second example allows you to run more 
complex code by providing the path to a script to run on the guest vm.  

Here's an example of a `sample_script.ps1` that will work with the environment 
variables we will set in our packer config; copy the contents into your own 
`sample_script.ps1` and provide the path to it in your packer config:

```
Write-Output("PACKER_BUILD_NAME is automatically set for you,)
Write-Output("or you can set it in your builder variables; )
Write-Output("the default for this builder is: " + $Env:PACKER_BUILD_NAME )
Write-Output("Remember that escaping variables in powershell requires backticks: )
Write-Output("for example, VAR1 from our config is " + $Env:VAR1 )
Write-Output("Likewise, VAR2 is " + $Env:VAR2 )
Write-Output("and VAR3 is " + $Env:VAR3 )
```

Next you need to create a packer config that will use this bootstrap file. See 
the example below, which contains examples of using source_ami_filter for 
windows in addition to the powershell and windows-restart provisioners:

```
{
  "variables": {
        "aws_access_key": "{{env `AWS_ACCESS_KEY_ID`}}",
        "aws_secret_key": "{{env `AWS_SECRET_ACCESS_KEY`}}",
        "region":         "us-east-1"
  },
  "builders": [
  {
    "type": "amazon-ebs",
    "access_key": "{{ user `aws_access_key` }}",
    "secret_key": "{{ user `aws_secret_key` }}",
    "region": "us-east-1",
    "instance_type": "m3.medium",
    "source_ami_filter": {
      "filters": {
        "virtualization-type": "hvm",
        "name": "*WindowsServer2012R2*",
        "root-device-type": "ebs"
      },
      "most_recent": true,
      "owners": "amazon"
    },    
    "ami_name": "packer-demo-{{timestamp}}",
    "user_data_file": "./bootstrap_win.txt",
    "communicator": "winrm",
    "winrm_username": "Administrator",
    "winrm_password": "SuperS3cr3t!"
  }],
  "provisioners": [
    {
      "type": "powershell",
      "environment_vars": ["DEVOPS_LIFE_IMPROVER=PACKER"],
      "inline": "Write-Output(\"HELLO NEW USER; WELCOME TO $Env:DEVOPS_LIFE_IMPROVER\")"
    },
    {
      "type": "windows-restart"
    },
    {
      "script": "./sample_script.ps1",
      "type": "powershell",
      "environment_vars": [
        "VAR1=A`$Dollar",
        "VAR2=A``Backtick",
        "VAR3=A`'SingleQuote"
      ]
    }
  ]
}
```

Then `packer build firstrun.json`

You should see output like this:

```
amazon-ebs output will be in this color.

==> amazon-ebs: Prevalidating AMI Name: packer-demo-1507234504
    amazon-ebs: Found Image ID: ami-d79776ad
==> amazon-ebs: Creating temporary keypair: packer_59d692c8-81f9-6a15-2502-0ca730980bed
==> amazon-ebs: Creating temporary security group for this instance: packer_59d692f0-dd01-6879-d8f8-7765327f5365
==> amazon-ebs: Authorizing access to port 5985 on the temporary security group...
==> amazon-ebs: Launching a source AWS instance...
==> amazon-ebs: Adding tags to source instance
    amazon-ebs: Adding tag: "Name": "Packer Builder"
    amazon-ebs: Instance ID: i-04467596029d0a2ff
==> amazon-ebs: Waiting for instance (i-04467596029d0a2ff) to become ready...
==> amazon-ebs: Skipping waiting for password since WinRM password set...
==> amazon-ebs: Waiting for WinRM to become available...
    amazon-ebs: WinRM connected.
==> amazon-ebs: Connected to WinRM!
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: /var/folders/8t/0yb5q0_x6mb2jldqq_vjn3lr0000gn/T/packer-powershell-provisioner079851514
    amazon-ebs: HELLO NEW USER; WELCOME TO PACKER
==> amazon-ebs: Restarting Machine
==> amazon-ebs: Waiting for machine to restart...
    amazon-ebs: WIN-164614OO21O restarted.
==> amazon-ebs: Machine successfully restarted, moving on
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: ./scripts/sample_script.ps1
    amazon-ebs: PACKER_BUILD_NAME is automatically set for you, or you can set it in your builder variables; the default for this builder is: amazon-ebs
    amazon-ebs: Remember that escaping variables in powershell requires backticks; for example VAR1 from our config is A$Dollar
    amazon-ebs: Likewise, VAR2 is A`Backtick
    amazon-ebs: and VAR3 is A'SingleQuote
==> amazon-ebs: Stopping the source instance...
    amazon-ebs: Stopping instance, attempt 1
==> amazon-ebs: Waiting for the instance to stop...
==> amazon-ebs: Creating the AMI: packer-demo-1507234504
    amazon-ebs: AMI: ami-2970b753
==> amazon-ebs: Waiting for AMI to become ready...
==> amazon-ebs: Terminating the source AWS instance...
==> amazon-ebs: Cleaning up any extra volumes...
==> amazon-ebs: No volumes to clean up, skipping
==> amazon-ebs: Deleting temporary security group...
==> amazon-ebs: Deleting temporary keypair...
Build 'amazon-ebs' finished.

==> Builds finished. The artifacts of successful builds are:
--> amazon-ebs: AMIs were created:
us-east-1: ami-2970b753
```

And if you navigate to your EC2 dashboard you should see your shiny new AMI.


[platforms]: /docs/builders/index.html
