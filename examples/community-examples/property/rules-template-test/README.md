# Terraform Use Cases for Akamai provider: SaaS and Workspaces
This demo includes the following use cases for Terarform with the Akamai provider:

* **[Akamai Snippets](#akamai-property-json-snippets)**. By using the Akamai Property Manager CLI a property can be broken down into several json snippet files, and inside these files use variables to replace certain values based on your variable definitions.
* **[SaaS](#saas)**. The Terraform for_each (introduced in [Terraform version 0.13.0](https://github.com/hashicorp/terraform/blob/v0.13/CHANGELOG.md)) function allows to iterate through variables/parameter sets resulting in deploying multiple configurations in a single plan/apply.
* **[Workspaces](#terraform-workspaces)**. Allows to reuse the *.tf and templates files in different environments (i.e. dev, qa, prod) keeping the SaaS functionality. Different state files will keep track of the different environments.

*Keyword(s):* terraform, akamai provider, automation, for_each, SaaS, workspace<br>

## Prerequisites
- [Akamai API Credentials](https://developer.akamai.com/getting-started/edgegrid) for creating propertie, hostnames and CP codes.
- [Akamai provider version 1.5.1](https://registry.terraform.io/providers/akamai/akamai/1.5.1). For older versions of the provider some of the fields are no longer supported. See the [change log](https://github.com/akamai/terraform-provider-akamai/blob/master/CHANGELOG.md) for the Akamai provider for more details.
- [Akamai Property Manager CLI](https://github.com/akamai/cli-property-manager)

## Akamai Property JSON Snippets
Use the Akamai Property Manager CLI to break down a property into several json files. Each json snippet will contain the rules and behaviors for each parent rule by default. The snippets (JSON templates) will be used in this example in the [SaaS](#saas) and [Workspaces](#terraform-workspaces) use cases, however here's a brief description on how they integrate with Terraform.

1. An existing property to use as baseline can be imported into snippets by submitting the following snippets command:

`$ akamai snippets import -p <PROPERTY-NAME>`

The resulting JSON snippets will be located under the folder PROPERTY-NAME/config-snippets. You can copy these files under a different folder where the TF configuration will reference to. In this example these have been moved to the Terraform project folder as config-snippets/.

2. In your main *.tf file use the data source ["akamai_property_rules_template"](https://registry.terraform.io/providers/akamai/akamai/latest/docs/data-sources/property_rules_template) which lets you configure a rule tree through the use of JSON template files (snippets). You can also keep your Property Manager CLI variable definition files and references (`“${env.<variableName>}"`)

This is an example use case where the `“${env.cpcode}"` and `“${env.origin}"` are replaced in your JSON snippets by the values defined below and the final rule tree is built:

```
data "akamai_property_rules_template" "rules" {
  template_file = abspath("${path.module}/config-snippets/main.json")
  variables {
    name="cpcode"
    value=some-cp-code
    type="number"
  }
  variables {
    name="origin"
    value="some.origin.com"
    type="string"
  }
}
```

## SaaS
The Terraform `for_each` function allows to iterate through variables/parameter sets resulting in deploying multiple configurations in a single plan/apply.
Steps to set it up assuming the property snippets (see above) have been created:

1. Create a variables.tf file which will contain the variable definitions. In this example we're defining the cpcode, origin, hostname and edge_hostname as variables inside the `properties' object.
```
variable "properties" {
    type = map(object({
        cpcode = string
        origin = string
        hostname = string
        edge_hostname = string
    }))
}
```
2. Create the terraform.tfvars which will contain the actual variable values. Here’s where all the parameters for the different configurations are added. Just like you would do in the environments/{env}/variables.json in the Akamai Pipeline CLI.
```
properties =  {
    "property1.name.goes.here" = {
        cpcode = "<cpcode#1>"
        origin = "origin1.name.goes.here"
        hostname = "hostname1.goes.here"
        edge_hostname = "edge-hostname1.goes.here"
    },
    "property2.name.goes.here" = {
        cpcode = "<cpcode#2>"
        origin = "origin2.name.goes.here"
        hostname = "hostname2.goes.here"
        edge_hostname = "edge-hostname2.goes.here"
    }
}
```
This example assumes the CP Codes are available before hand. However in a real scenario you'll want to create the new CP Codes on the fly as well.
**IMPORTANT**: The property name is actually the key in the snippet above. Just like you would have in a JSON file.
 
3. In the main akamai.tf then use the `for_each` function to create the iteration effect. For example for the property creation and edge hostname creation blocks. This could be done for the CP code and Property Activation too.

```
resource "akamai_edge_hostname" "new-edge-hostname" {
 
for_each = var.properties
 
product_id  = "prd_SPM"
contract_id = data.akamai_contract.contract.id
group_id = data.akamai_group.group.id
ip_behavior = "IPV6_COMPLIANCE"
edge_hostname = each.value.edge_hostname
certificate = <CERT_ENROLLMENT_ID>
}
 

resource "akamai_property" "new-property" {
 
for_each = var.properties
 
name = each.key
contract_id = data.akamai_contract.contract.id
group_id = data.akamai_group.group.id
product_id = "prd_SPM"
rule_format = "latest"
hostnames {
  cname_from = each.value.hostname
  cname_to = akamai_edge_hostname.new-edge-hostname[each.key].edge_hostname
  cert_provisioning_type = "CPS_MANAGED"
}
rules = data.akamai_property_rules_template.rules[each.key].json
}
 

data "akamai_property_rules_template" "rules" {
 
  for_each = var.properties
 
  template_file = abspath("${path.module}/config-snippets/main.json")
  variables {
    name="cpcode"
    value=each.value.cpcode
    type="number"
  }
  variables {
    name="origin"
    value=each.value.origin
    type="string"
  }
}
```

For the resource “akamai_property” observe the property name is just the key (`each.key`). And that the edge hostname (`cname_to`) now is built based on the parameters for the current iteration (`each.key`). The same goes for the `“rules”` setting.
Then under each block (`akamai_property_rules_template` and `akamai_edge_hostname`) the actual values are referenced, for example for the origin the name would be set as `each.value.origin` and so on.
 
By using the SaaS approach you can see how you can also work with different environments (dev, qa, prod) just by manipulating variables. However there’s an alternative: Workspaces.


## Terraform Workspaces
Workspaces allow to reuse the same snippets/templates and *.tf files, and create different state files based on your environments.
In this repository you’ll see the folder environments/ which will contain different tfvars files which correspond to each one of the environments to set up:

* environment/dev.tfvars
* environment/prod.tfvars
 
To set up the “dev” environment run the following command
 
`$ terraform workspace new dev`
 
You can check the available workspaces and the one you’re working on:

``` 
$ terraform workspace list
  default
* dev
```

Finally, to plan and apply you can do it this way:
 
`$ terraform plan -var-file environments/$(terraform workspace show).tfvars`
 
`$ terraform apply -var-file environments/$(terraform workspace show).tfvars`
 
## Debugging
For debugging you can enable the logging level: OFF, TRACE, DEBUG, INFO, WARN or ERROR
 
`$ export TF_LOG=TRACE`
 
TRACE will give the Akamai outputs.

## Future Improvements
* Add CP code creation code.
* Add property activation code.
* Maybe integrate with some cloud deployment which could serve as an origin.

## More details on Terraform and Akamai Provider
- [Terraform](https://www.terraform.io/)
- [Akamai provider](https://registry.terraform.io/providers/akamai/akamai/latest)
- [Akamai Terraform Examples](https://github.com/akamai/terraform-provider-akamai/tree/master/examples)