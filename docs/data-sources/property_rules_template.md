---
layout: "akamai"
page_title: "Akamai: akamai_property_rules_template"
subcategory: "Property Provisioning"
description: |-
 Property Rules Template
---

# akamai_property_rules_template

The `akamai_property_rules_template` data source lets you configure a rule tree through the use of JSON template files. A rule tree is a nested block of property
rules in JSON format that include match criteria and behaviors.

With this data source you define the location of the JSON template files and provide information about any user-defined variables included within the templates.

The template format used in this data source matches those used in the [Property Manager CLI](https://learn.akamai.com/en-us/learn_akamai/getting_started_with_akamai_developers/developer_tools/getstartedpmcli.html#addanewsnippet).

You can pass user-defined variables by supplying either:

* paths to `variableDefinitions.json` and `variables.json` with syntax used in Property Manager CLI, or
* a set of Terraform variables.

## Referencing sub-files from a template
You can split each template out into a series of smaller template files. To add
them to this data source, you need to include them in the currently loaded file,
which corresponds to the value in the `template_file` argument.  For example, to
include `example-file.json` from the `property-snippets` directory, use this syntax
including the quotes: `"#include:example-file.json"`.  Make sure the `property-snippets` folder contains only `.json` files.
All files are resolved in relation to the directory that contains the starting template file.

## Inserting variables in a template
You can also add variables to a template by using a string like `“${env.<variableName>}"`. You'll need the quotes here too.  
These variables follow the format used in the [Property Manager CLI](https://github.com/akamai/cli-property-manager#update-the-variabledefinitions-file).  They differ from Terraform variables which should resolve normally.

## Example usage: variables

This first example shows two variables passed in data source definition:

```hcl
data "akamai_property_rules_template" "akarules" {
  template_file = abspath("${path.root}/rules/main.json")
  variables {
    name = "enabled"
    value = "true"
    type = "bool"
  }
  variables {
    name = "name"
    value = "default"
    type = "string"
  }
}
```

In this second example, the variables defined refer to files shared with a [Property Manager CLI pipeline](https://github.com/akamai/cli-property-manager#akamai-pipeline-workflow):

```hcl
data "akamai_property_rules_template" "akarules" {
  template_file = abspath("${path.root}/property-snippets/main.json")
  var_definition_file = abspath("${path.root}/environments/variableDefinitions.json")
  var_values_file = abspath("${path.root}/environments/dev.example.com/variables.json")
}
```

### Example usage: template files

Here's an example of what a JSON-based template file with its nested templates might look like:

property-snippets/main.json:
```json
{
  "rules": {
    "name": "default",
    "children": [
      "#include:Default_GKE_Origin.json",
      "#include:Origin_Selection.json",
      "#include:Performance.json",
      "#include:Offload.json",
      "#include:Restrict_Method.json",
      "#include:Set_PMUSER_VARS.json",
      "#include:Modify_cache_key.json"
    ],
    "behaviors": "#include:behaviors_default.json",
    "options": {
      "is_secure": "${env.secure}"
    },
    "variables": "#include:akamai_variables.json",
    "comments": "The behaviors in the Default Rule apply to all requests for the property hostnames unless another rule overrides the Default Rule settings."
  }
}
```

You can then define a Terraform configuration file like this, which pulls in the `main.json` file above and uses it with a property:

```hcl
data "akamai_property_rules_template" "example" {
  template_file = abspath("${path.root}/property-snippets/main.json")
  variables {
      name = "secure"
      value = "false"
      type = "bool"
  }
  variables {
      name = "caching_ttl"
      value = "3d"
      type = "string"
  }
}

resource "akamai_property" "example" {
    name = "dev.example.com"
    contract_id = var.contractid
    group_id    = var.groupid
    hostnames = {
      "example.org" = "example.org.edgesuite.net"
      "www.example.org" = "example.org.edgesuite.net"
      "sub.example.org" = "sub.example.org.edgesuite.net"
    }
    rule_format = "v2020-03-04"
    rules       = data.akamai_property_rules_template.example.json
}
```

## Argument reference

* `template_file` - (Required) The absolute path to your top-level JSON template file. The top-level template combines smaller, nested JSON templates to form your property rule tree.
* `variables` - (Optional) A definition of a variable. Variables aren't required and you can use multiple ones if needed. This argument conflicts with the `variable_definition_file` and `variable_values_file` arguments. A `variables` block includes:
    * `name` - The name of the variable used in template.
    * `type` - The type of variable: `string`, `number`, `bool`, `jsonArray` or `jsonBlock`.
    * `value` - The value of the variable passed as a string.
* `variable_definition_file` - (Optional) The absolute path to the file containing variable definitions and defaults. This file follows the syntax used in the [Property Manager CLI](https://github.com/akamai/cli-property-manager). This argument is required if you set `variable_values_file` and conflicts with `variables`.
* `variable_values_file` - (Optional) The absolute path to the file containing variable values. This file follows the syntax used in the Property Manager CLI. This argument is required if you set `variable_definition_file` and conflicts with `variables`.

## Attributes reference

This data source returns this attribute:

* `json` - The fully expanded template with variables and all nested templates resolved.
