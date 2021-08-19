---
layout: "akamai"
page_title: "Akamai: Site Shield"
subcategory: "Site Shield Maps"
description: |-
 Site Shield
---

# akamai_akamai_siteshield_map

Use the `akamai_akamai_siteshield_map` data source to retrieve information about the Site Shield maps, filtered by map ID. The information available is described
[here](https://developer.akamai.com/api/cloud_security/site_shield/v1.html#getamap). 

## Example Usage

Basic usage:

```hcl
provider "akamai" {
  edgerc = "~/.edgerc"
}

data "akamai_siteshield_map" "siteshield" {
   map_id = 1234
}

output "siteshield_current_cidrs" {
  value = data.akamai_siteshield_map.siteshield.current_cidrs
}

output "siteshield_proposed_cidrs" {
  value = data.akamai_siteshield_map.siteshield.proposed_cidrs
}

output "siteshield_rule_name" {
  value = data.akamai_siteshield_map.siteshield.rule_name
}

output "siteshield_acknowledged" {
  value = data.akamai_siteshield_map.siteshield.acknowledged
}

```

## Argument Reference

* `map_id` - (Required) The map ID of a specific Site Shield map to retrieve. 

The following arguments are supported:

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `current_cidrs` - A list of current CIDRs configured for the specified SS map.

* `proposed_cidrs` - A list of proposed (new) CIDRs configured for the specified SS map.

* `rule_name` - A map rule name available shown in properties.

* `acknowledged` - A boolean of the aknowledgement state of the map.

