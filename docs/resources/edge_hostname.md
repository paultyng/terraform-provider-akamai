---
layout: "akamai"
page_title: "Akamai: edge hostname"
subcategory: "Property Provisioning"
description: |-
  Edge Hostname
---

# akamai_edge_hostname

~> **Note** Version 1.0.0 of the Akamai Terraform Provider is now available for the Property Provisioning module. To upgrade to the new version, you have to update this resource. See the [upgrade guide](../guides/1.0_migration.md) for details.

The `akamai_edge_hostname` resource lets you configure a secure edge hostname. Your
edge hostname determines how requests for your site, app, or content are mapped to
Akamai edge servers.

An edge hostname is the CNAME target you use when directing your end user traffic to
Akamai. Each hostname assigned to a property has a corresponding edge hostname.

Akamai supports three types of edge hostnames, depending on the level of security
you need for your traffic: Standard TLS, Enhanced TLS, and Shared Certificate. When
entering the `edge_hostname` attribute, you need to include a specific domain suffix
for your edge hostname type:

| Edge hostname type | Domain suffix |
|------|-------|
| Enhanced TLS | edgekey.net |
| Standard TLS | edgesuite.net |
| Shared Cert | akamaized.net |

For example, if you use Standard TLS and have `www.example.com` as a hostname, your edge hostname would be `www.example.com.edgesuite.net`. If you wanted to use Enhanced TLS with the same hostname, your edge hostname would be `www.example.com.edgekey.net`. See the [Property Manager API (PAPI)](https://developer.akamai.com/api/core_features/property_manager/v1.html#createedgehostnames) for more information.

## Example usage

Basic usage:

```hcl
resource "akamai_edge_hostname" "terraform-demo" {
    product_id  = "prd_Object_Delivery"
    contract_id = "ctr_1-AB123"
    group_id    = "grp_123"
    edge_hostname = "www.example.org.edgesuite.net"
}
```

## Argument reference

This resource supports these arguments:

* `name` - (Required) The name of the edge hostname.
* `contract_id` - (Required) A contract's unique ID, including the `ctr_` prefix.
* `group_id` - (Required) A group's unique ID, including the `grp_` prefix.
* `product_id` - (Required) A product's unique ID, including the `prd_` prefix.
* `edge_hostname` - (Required) One or more edge hostnames. The number of edge hostnames must be less than or equal to the number of public hostnames.
* `certificate` - (Optional) Required only when creating an Enhanced TLS edge hostname. This argument sets the certificate enrollment ID. Edge hostnames for Enhanced TLS end in `edgekey.net`. You can retrieve this ID from the [Certificate Provisioning Service CLI](https://github.com/akamai/cli-cps) .
* `ip_behavior` - (Required) Which version of the IP protocol to use: `IPV4` for version 4 only, `IPV6_PERFORMANCE` for version 6 only, or `IPV6_COMPLIANCE` for both 4 and 6.

### Deprecated arguments

* `contract` - (Deprecated) Replaced by `contract_id`. Maintained for legacy purposes.
* `group` - (Deprecated) Replaced by `group_id`. Maintained for legacy purposes.
* `product` - (Deprecated) Replaced by `product_id`. Maintained for legacy purposes.

## Attributes reference

This resource returns this attribute:

* `ip_behavior` - Returns the IP protocol the hostname will use, either `IPV4` for version 4, IPV6_PERFORMANCE` for version 6, or `IPV6_COMPLIANCE` for both.

## Import

Basic Usage:

```hcl
resource "akamai_edge_hostname" "example" {
    # (resource arguments)
  }
```

You can import Akamai edge hostnames using a comma-delimited string of edge
hostname, contract ID, and group ID. You have to enter the values in this order:

 `edge_hostname, contract_id, group_id`

For example:

```shell
$ terraform import akamai_edge_hostname.example ehn_123,ctr_1-AB123,grp_123
```
