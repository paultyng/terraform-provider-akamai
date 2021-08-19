---
layout: "akamai"
page_title: "Module: Site Shield Maps"
description: |-
  Site Shield Maps module for the Akamai Terraform Provider
---

# Site Shield Maps Module Guide

The Akamai Site Shield Maps provider for Terraform gives you the ability to automate the retrieval of Site Shield maps used in various Akamai products. For customers who are already using the Akamai Network, Site Shield provides an additional layer of protection that helps prevent attackers from bypassing cloud-based protections to target the application origin. Site Shield cloaks websites and applications from the public Internet and restricts clients from directly accessing the origin. It is designed to complement the existing network infrastructure as well as advanced cloud security technologies available on the globally-distributed Akamai Intelligent Platform to mitigate the risks associated with network and application-layer threats that directly target the origin infrastructure. For more information about Site Shield Maps, see the [API documentation](https://developer.akamai.com/api/cloud_security/site_shield/v1.html)

## Configure the Terraform Provider

Set up your .edgerc credential files as described in [Get Started with Akamai APIs](https://developer.akamai.com/api/getting-started), and include read-write permissions for the Network Lists API. 

1. Create a new folder called `terraform`
1. Inside the new folder, create a new file called `akamai.tf`.
1. Add the provider configuration to your `akamai.tf` file:

```hcl
provider "akamai" {
	edgerc = "~/.edgerc"
	config_section = "siteshield"
}
```

## Prerequisites

Review [Get Started with APIs](https://learn.akamai.com/en-us/learn_akamai/getting_started_with_akamai_developers/developer_tools/getstartedapis.html) for details on how to set up client tokens to access any Akamai API. These tokens appear as custom hostnames that look like this: https://akzz-XXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXX.luna.akamaiapis.net.

To enable this API, choose the API service named Network Lists, and set the access level to READ-WRITE.

## Retrieving Site Shield Map Information

You can obtain a list of all network lists available for an authenticated user belonging to a group using the [`akamai_siteshield_map`](../data-sources/siteshield_map.md) data source and its `output_text` attribute. Add the following to your `akamai.tf` file:

```hcl
data "akamai_siteshield_map" "siteshield" {
   map_id = 1234
}

output "siteshield_proposed_cidrs" {
  value = data.akamai_siteshield_map.siteshield.proposed_cidrs
}
```

Once you have saved the file, switch to the terminal and initialize Terraform using the command:

```bash
$ terraform init
```

This command will install the latest version of the Akamai provider, as well as any other providers necessary. To update the Akamai provider version after a new release, simply run `terraform init` again.

## Test Your Configuration

To test your configuration, use `terraform plan`:

```bash
$ terraform plan
```

This command will make Terraform create a plan for the work it will do based on the configuration file. This will not actually make any changes and is safe to run as many times as you like.

## Apply Changes

To actually display the configuration information, or to create or modify resources as described further in this guide, we need to instruct Terraform to `apply` the changes outlined in the plan. To do this, in the terminal, run the command:

```bash
$ terraform apply
```
