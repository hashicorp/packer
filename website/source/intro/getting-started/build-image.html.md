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
Our first image will be an [Amazon EC2 AMI](https://aws.amazon.com/ec2/).
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

With a properly validated template, it is time to build your first image. This
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
they're built, it is up to you to launch or destroy them as you see fit.

After running the above example, your AWS account now has an AMI associated
with it. AMIs are stored in S3 by Amazon, so unless you want to be charged
about &#36;0.01 per month, you'll probably want to remove it. Remove the AMI by
first deregistering it on the [AWS AMI management
page](https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Images). Next,
delete the associated snapshot on the [AWS snapshot management
page](https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Snapshots).

Congratulations! You've just built your first image with Packer. Although the
image was pretty useless in this case (nothing was changed about it), this page
should've given you a general idea of how Packer works, what templates are and
how to validate and build templates into machine images.

## Some more examples:

### Another GNU/Linux Example, with provisioners:
Create a file named `welcome.txt` and add the following:

```
WELCOME TO PACKER!
```

Create a file named `example.sh` and add the following:


```bash
#!/bin/bash
echo "hello"
```

Set your access key and id as environment variables, so we don't need to pass
them in through the command line:

```
export AWS_ACCESS_KEY_ID=MYACCESSKEYID
export AWS_SECRET_ACCESS_KEY=MYSECRETACCESSKEY
```

Now save the following text in a file named `firstrun.json`:

```json
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
it might look something like this: `"source_ami": "ami-fce3c696"`.

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

### A Windows Example

As with the GNU/Linux example above, should you decide to follow along and
build an AMI from the example template, provided you qualify for free tier
usage, you should not be charged for actually building the AMI.
However, please note that you will be charged for storage of the snapshot
associated with any AMI that you create.
If you wish to avoid further charges, follow the steps in the [Managing the
Image](/intro/getting-started/build-image.html#managing-the-image) section
above to deregister the created AMI and delete the associated snapshot once
you're done.

Again, in this example, we are making use of an existing AMI available from
the Amazon marketplace as the *source* or starting point for building our
own AMI. In brief, Packer will spin up the source AMI, connect to it and then
run whatever commands or scripts we've configured in our build template to
customize the image. Finally, when all is done, Packer will wrap the whole
customized package up into a brand new AMI that will be available from the
[AWS AMI management page](
https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Images). Any
instances we subsequently create from this AMI will have all of our
customizations baked in. This is the core benefit we are looking to
achieve from using the [Amazon EBS builder](/docs/builders/amazon-ebs.html)
in this example.

Now, all this sounds simple enough right? Well, actually it turns out we
need to put in just a *bit* more effort to get things working as we'd like...

Here's the issue: Out of the box, the instance created from our source AMI
is not configured to allow Packer to connect to it. So how do we fix it so
that Packer can connect in and customize our instance?

Well, it turns out that Amazon provides a mechanism that allows us to run a
set of *pre-supplied* commands within the instance shortly after the instance
starts. Even better, Packer is aware of this mechanism. This gives us the
ability to supply Packer with the commands required to configure the instance
for a remote connection *in advance*. Once the commands are run, Packer
will be able to connect directly in to the instance and make the
customizations we need.

Here's a basic example of a file that will configure the instance to allow
Packer to connect in over WinRM. As you will see, we will tell Packer about
our intentions by referencing this file and the commands within it from
within the `"builders"` section of our
[build template](/docs/templates/index.html) that we will create later.

Note the `<powershell>` and `</powershell>` tags at the top and bottom of
the file. These tags tell Amazon we'd like to run the enclosed code with
PowerShell. You can also use `<script></script>` tags to enclose any commands
that you would normally run in a Command Prompt window. See
[Running Commands on Your Windows Instance at Launch](
http://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/ec2-windows-user-data.html)
for more info about what's going on behind the scenes here.

```powershell
<powershell>
# Set administrator password
net user Administrator SuperS3cr3t!
wmic useraccount where "name='Administrator'" set PasswordExpires=FALSE

# First, make sure WinRM can't be connected to
netsh advfirewall firewall set rule name="Windows Remote Management (HTTP-In)" new enable=yes action=block

# Delete any existing WinRM listeners
winrm delete winrm/config/listener?Address=*+Transport=HTTP  2>$Null
winrm delete winrm/config/listener?Address=*+Transport=HTTPS 2>$Null

# Create a new WinRM listener and configure
winrm create winrm/config/listener?Address=*+Transport=HTTP
winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="0"}'
winrm set winrm/config '@{MaxTimeoutms="7200000"}'
winrm set winrm/config/service '@{AllowUnencrypted="true"}'
winrm set winrm/config/service '@{MaxConcurrentOperationsPerUser="12000"}'
winrm set winrm/config/service/auth '@{Basic="true"}'
winrm set winrm/config/client/auth '@{Basic="true"}'

# Configure UAC to allow privilege elevation in remote shells
$Key = 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System'
$Setting = 'LocalAccountTokenFilterPolicy'
Set-ItemProperty -Path $Key -Name $Setting -Value 1 -Force

# Configure and restart the WinRM Service; Enable the required firewall exception
Stop-Service -Name WinRM
Set-Service -Name WinRM -StartupType Automatic
netsh advfirewall firewall set rule name="Windows Remote Management (HTTP-In)" new action=allow localip=any remoteip=any
Start-Service -Name WinRM
</powershell>
```

Save the above code in a file named `bootstrap_win.txt`.

-> **A quick aside/warning:**<br />
Windows administrators in the know might be wondering why we haven't simply
used a `winrm quickconfig -q` command in the script above, as this would
*automatically* set up all of the required elements necessary for connecting
over WinRM. Why all the extra effort to configure things manually?<br />
Well, long and short, use of the `winrm quickconfig -q` command can sometimes
cause the Packer build to fail shortly after the WinRM connection is
established. How?<br />
1. Among other things, as well as setting up the listener for WinRM, the
quickconfig command also configures the firewall to allow management messages
to be sent over HTTP.<br />
2. This undoes the previous command in the script that configured the
firewall to prevent this access.<br />
3. The upshot is that the system is configured and ready to accept WinRM
connections earlier than intended.<br />
4. If Packer establishes its WinRM connection immediately after execution of
the 'winrm quickconfig -q' command, the later commands within the script that
restart the WinRM service will unceremoniously pull the rug out from under
the connection.<br />
5. While Packer does *a lot* to ensure the stability of its connection in to
your instance, this sort of abuse can prove to be too much and *may* cause
your Packer build to stall irrecoverably or fail!

Now we've got the business of getting Packer connected to our instance
taken care of, let's get on with the *real* reason we're doing all this,
which is actually configuring and customizing the instance. Again, we do this
with [Provisioners](/docs/provisioners/index.html).

The example config below shows the two different ways of using the [PowerShell
provisioner](/docs/provisioners/powershell.html): `inline` and `script`.
The first example, `inline`, allows you to provide short snippets of code, and
will create the script file for you.  The second example allows you to run more
complex code by providing the path to a script to run on the guest VM.

Here's an example of a `sample_script.ps1` that will work with the environment
variables we will set in our build template; copy the contents into your own
`sample_script.ps1` and provide the path to it in your build template:

```powershell
Write-Host "PACKER_BUILD_NAME is an env var Packer automatically sets for you."
Write-Host "...or you can set it in your builder variables."
Write-Host "The default for this builder is:" $Env:PACKER_BUILD_NAME

Write-Host "The PowerShell provisioner will automatically escape characters"
Write-Host "considered special to PowerShell when it encounters them in"
Write-Host "your environment variables or in the PowerShell elevated"
Write-Host "username/password fields."
Write-Host "For example, VAR1 from our config is:" $Env:VAR1
Write-Host "Likewise, VAR2 is:" $Env:VAR2
Write-Host "VAR3 is:" $Env:VAR3
Write-Host "Finally, VAR4 is:" $Env:VAR4
Write-Host "None of the special characters needed escaping in the template"
```

Finally, we need to create the actual [build template](
/docs/templates/index.html).
Remember, this template is the core configuration file that Packer uses to
understand what you want to build, and how you want to build it.

As mentioned earlier, the specific builder we are using in this example
is the [Amazon EBS builder](/docs/builders/amazon-ebs.html).
The template below demonstrates use of the [`source_ami_filter`](
/docs/builders/amazon-ebs.html#source_ami_filter) configuration option
available within the builder for automatically selecting the *latest*
suitable source Windows AMI provided by Amazon.
We also use the `user_data_file` configuration option provided by the builder
to reference the bootstrap file we created earlier. As you will recall, our
bootstrap file contained all the commands we needed to supply in advance of
actually spinning up the instance, so that later on, our instance is
configured to allow Packer to connect in to it.

The `"provisioners"` section of the template demonstrates use of the
[powershell](/docs/provisioners/powershell.html) and
[windows-restart](/docs/provisioners/windows-restart.html) provisioners to
customize and control the build process:

```json
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
      "region": "{{ user `region` }}",
      "instance_type": "t2.micro",
      "source_ami_filter": {
        "filters": {
          "virtualization-type": "hvm",
          "name": "*Windows_Server-2012-R2*English-64Bit-Base*",
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
    }
  ],
  "provisioners": [
    {
      "type": "powershell",
      "environment_vars": ["DEVOPS_LIFE_IMPROVER=PACKER"],
      "inline": [
        "Write-Host \"HELLO NEW USER; WELCOME TO $Env:DEVOPS_LIFE_IMPROVER\"",
        "Write-Host \"You need to use backtick escapes when using\"",
        "Write-Host \"characters such as DOLLAR`$ directly in a command\"",
        "Write-Host \"or in your own scripts.\""
      ]
    },
    {
      "type": "windows-restart"
    },
    {
      "script": "./sample_script.ps1",
      "type": "powershell",
      "environment_vars": [
        "VAR1=A$Dollar",
        "VAR2=A`Backtick",
        "VAR3=A'SingleQuote",
        "VAR4=A\"DoubleQuote"
      ]
    }
  ]
}
```

Save the build template as `firstrun.json`.

Next we need to set things up so that Packer is able to access and use our
AWS account. Set your access key and id as environment variables, so we
don't need to pass them in through the command line:

```
export AWS_ACCESS_KEY_ID=MYACCESSKEYID
export AWS_SECRET_ACCESS_KEY=MYSECRETACCESSKEY
```

Finally, we can create our new AMI by running `packer build firstrun.json`

You should see output like this:

```
amazon-ebs output will be in this color.

==> amazon-ebs: Prevalidating AMI Name: packer-demo-1518111383
    amazon-ebs: Found Image ID: ami-013e197b
==> amazon-ebs: Creating temporary keypair: packer_5a7c8a97-f27f-6708-cc3c-6ab9b4688b13
==> amazon-ebs: Creating temporary security group for this instance: packer_5a7c8ab5-444c-13f2-0aa1-18d124cdb975
==> amazon-ebs: Authorizing access to port 5985 from 0.0.0.0/0 in the temporary security group...
==> amazon-ebs: Launching a source AWS instance...
==> amazon-ebs: Adding tags to source instance
    amazon-ebs: Adding tag: "Name": "Packer Builder"
    amazon-ebs: Instance ID: i-0c8c808a3b945782a
==> amazon-ebs: Waiting for instance (i-0c8c808a3b945782a) to become ready...
==> amazon-ebs: Skipping waiting for password since WinRM password set...
==> amazon-ebs: Waiting for WinRM to become available...
    amazon-ebs: WinRM connected.
==> amazon-ebs: Connected to WinRM!
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: /var/folders/15/d0f7gdg13rnd1cxp7tgmr55c0000gn/T/packer-powershell-provisioner943573503
    amazon-ebs: HELLO NEW USER; WELCOME TO PACKER
    amazon-ebs: You need to use backtick escapes when using
    amazon-ebs: characters such as DOLLAR$ directly in a command
    amazon-ebs: or in your own scripts.
==> amazon-ebs: Restarting Machine
==> amazon-ebs: Waiting for machine to restart...
    amazon-ebs: WIN-NI8N45RPJ23 restarted.
==> amazon-ebs: Machine successfully restarted, moving on
==> amazon-ebs: Provisioning with Powershell...
==> amazon-ebs: Provisioning with powershell script: ./sample_script.ps1
    amazon-ebs: PACKER_BUILD_NAME is an env var Packer automatically sets for you.
    amazon-ebs: ...or you can set it in your builder variables.
    amazon-ebs: The default for this builder is: amazon-ebs
    amazon-ebs: The PowerShell provisioner will automatically escape characters
    amazon-ebs: considered special to PowerShell when it encounters them in
    amazon-ebs: your environment variables or in the PowerShell elevated
    amazon-ebs: username/password fields.
    amazon-ebs: For example, VAR1 from our config is: A$Dollar
    amazon-ebs: Likewise, VAR2 is: A`Backtick
    amazon-ebs: VAR3 is: A'SingleQuote
    amazon-ebs: Finally, VAR4 is: A"DoubleQuote
    amazon-ebs: None of the special characters needed escaping in the template
==> amazon-ebs: Stopping the source instance...
    amazon-ebs: Stopping instance, attempt 1
==> amazon-ebs: Waiting for the instance to stop...
==> amazon-ebs: Creating the AMI: packer-demo-1518111383
    amazon-ebs: AMI: ami-f0060c8a
==> amazon-ebs: Waiting for AMI to become ready...
==> amazon-ebs: Terminating the source AWS instance...
==> amazon-ebs: Cleaning up any extra volumes...
==> amazon-ebs: No volumes to clean up, skipping
==> amazon-ebs: Deleting temporary security group...
==> amazon-ebs: Deleting temporary keypair...
Build 'amazon-ebs' finished.

==> Builds finished. The artifacts of successful builds are:
--> amazon-ebs: AMIs were created:
us-east-1: ami-f0060c8a
```

And if you navigate to your EC2 dashboard you should see your shiny new AMI
listed in the main window of the Images -> AMIs section.

Why stop there though?

As you'll see, with one simple change to the template above, it's
just as easy to create your own Windows 2008 or Windows 2016 AMIs. Just
set the value for the name field within `source_ami_filter` as required:

For Windows 2008 SP2:

```
          "name": "*Windows_Server-2008-SP2*English-64Bit-Base*",
```

For Windows 2016:

```
          "name": "*Windows_Server-2016-English-Full-Base*",
```

The bootstrapping and sample provisioning should work the same across all
Windows server versions.

[platforms]: /docs/builders/index.html
