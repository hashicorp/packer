---
description: |
    The Packer Amazon Import post-processor takes an OVA artifact from the VMware builder and
    imports it to an AMI available to Amazon Web Services EC2.
layout: docs
page_title: 'Amazon Import Post-Processor'
...

# Amazon Import Post-Processor

Type: `amazon-import`

The Packer Amazon Import post-processor takes an OVA artifact from the VMware builder and imports it to an AMI available to Amazon Web Services EC2.

\~&gt; This post-processor is for advanced users. Please ensure you read the ["prerequisites for import"](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/VMImportPrerequisites.html) before using this post-processor. You are strongly recommended to understand what behaviour is expected from an AMI before using this post-processor.

## How Does it Work?

The import process operates by copying the OVA to an S3 bucket, and calling an import task in EC2 on the OVA file. Once completed, an AMI is returned containing the converted virtual machine.

The import process itself run by AWS includes modifications to the image uploaded, to allow it to boot and operate in the AWS EC2 environment. However, not all modifications required to make the machine run well in EC2 are performed. Take care around console output from the machine, as debugging can be very difficult without it.

Further information about the import process can be found in AWS's ["EC2 Import/Export Instance documentation"](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instances_of_your_vm.html).

## Configuration

There are some configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

-   `s3_bucket` (string) - The name of the bucket where the OVA file will be copied to for import.

-   `s3_key` (string) - The name of the key where the OVA file will be copied to for import.

Optional:

-   `skip_clean` (boolean) - Whether we should skip removing the OVA file uploaded to S3 after the import process has completed. "true" means that we should leave it in the S3 bucket, "false" means to clean it out. Defaults to "false".

