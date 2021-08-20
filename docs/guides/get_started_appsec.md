---
layout: "akamai"
page_title: "Module: Application Security"
description: |-
   Application Security module for the Akamai Terraform Provider
---

# Application Security Guide 

Application Security (appsec) in the Akamai Terraform provider (provider) enables application 
security configurations including the following: 
* custom rules.
* match targets.
* other application security resources that operate within the Cloud.

This Guide is for developers who:
* are interested in implementing or updating an integration of Akamai functionality with Terraform.
* already have some familiarity with Akamai.
 
## Prerequisites

~> **Note** The Application Security subprovider is currently in beta. If you’re currently using the Application Security module,
the latest Akamai Terraform Provider release, v1.6.0, includes breaking changes that require you to update your existing Terraform
configuration.

To manage Application Security resources, you need to obtain information regarding your 
existing security implementation, including the following information:

* **Configuration ID**: The ID of the specific security configuration under which the resources are defined.

### Retrieve existing security configuration information

You can obtain the name and ID of the existing security configurations by using the 
[`akamai_appsec_configuration`](../data-sources/appsec_configuration.md) data source. 
Using it without parameters outputs information about all security configurations associated with your account. 

Add the following to your `akamai.tf` file:

```hcl
data "akamai_appsec_configuration" "configurations" {
}

output "configuration_list" {
  value = data.akamai_appsec_configuration.configurations.output_text
}
```

Save the resulting text file, and then use terminal to initialize Terraform with the command:

```bash
$ terraform init
```

This installs the latest version of the Akamai provider, along with any other providers necessary. 

When you need to obtain an update of Akamai provider, run `terraform init` again.

## Configure the Provider

Set up your .edgerc credential files as described in
[Get Started with Akamai APIs](https://developer.akamai.com/api/getting-started),
and include read-write permissions for the Application Security API. 

1. Create a new folder called `terraform`
1. Inside the new folder, create a new file called `akamai.tf`.
1. Add the provider configuration to your `akamai.tf` file:

```hcl
provider "akamai" {
  edgerc = "~/.edgerc"
  config_section = "appsec"
}
```

## Test Your Configuration

To test your configuration, use `terraform plan`:

```bash
$ terraform plan
```

This command causes Terraform to create a plan for the work it will do, based on the configuration (*.tf) files
in the current directory. This does *not* actually make any changes and is safe to run multiple times.

## Apply Changes

To display existing configuration information, or to create or modify resources as described later in this guide,
tell Terraform to `apply` the changes outlined in the plan by running the command:

```bash
$ terraform apply
```

Given the configuration file above, Terraform will respond with a formatted list of all existing security
configurations in your account, giving their names and IDs (`config_id`), the most recently created version,
 and the version currently active in staging and production, if any.

When you have identified the desired security configuration by name, you can load that specific configuration
into Terraform's state. 

To load a specific configuration:
1. Identify the desired security configuration by name, 
1. Edit your `akamai.tf` file to add the desired `name` parameter to the `akamai_appsec_configuration` data block.
1. Change the `output` block so that it gives just the `config_id` attribute of the configuration. 

After these changes, the section of your file below the initial `provider` block looks like the following example:

```hcl
data "akamai_appsec_configuration" "configuration" {
  name = "Example"
}

output "ID" {
  value = data.akamai_appsec_configuration.configuration.config_id
}
```

After running `terraform apply` on this file, the terminal displays the specified `output` name and value, for example:

```hcl
ID = 12345
```

## Specify a Configuration to Display

The provider's [`akamai_appsec_export_configuration`](../data-sources/appsec_export_configuration.md) data source
can display complete information about any configuration that you specify, including attributes like custom rules,
and selected hostnames. 

To show custom rule and selected hostname data for your most recent configuration, add the following blocks to
your `akamai.tf` file:

```hcl
data "akamai_appsec_export_configuration" "export" {
  config_id = data.akamai_appsec_configuration.configuration.config_id
  version = data.akamai_appsec_configuration.configuration.latest_version
  search = [
  "customRules",
  "selectedHosts"
  ]
}

output "exported_configuration_text" {
  value = data.akamai_appsec_export_configuration.export.output_text
}
```

NOTE: You can also specify other kinds of data for export using any of the following fields in the `search` list:

 * attackGroups
 * customDenyList
 * customRules
 * matchTargets
 * ratePolicies
 * reputationProfiles
 * rules
 * securityPolicies
 * selectableHosts
 * selectedHosts

Save the file and run `terraform apply` to see a formatted display of the selected data.

## Add a Hostname to the `selectedHosts` List

You can modify the list of hosts protected by a specific security configuration using 
the [`akamai_appsec_selected_hostnames`](../data-sources/appsec_selected_hostnames.md) resource. 
Add the following resource block to your `akamai.tf` file, replacing `example.com` with a hostname 
from the list reported in the `data_akamai_appsec_export_configuration` data source example above:

```hcl
resource "akamai_appsec_selected_hostnames" "selected_hostnames_append" {
  config_id = data.akamai_appsec_configuration.configuration.config_id
  hostnames = [ "example.com" ]
  mode = "APPEND"
}

output "selected_hostnames_appended" {
  value = akamai_appsec_selected_hostnames.selected_hostnames_append.hostnames
}
```

When you specify changes to an existing resource as described here, or when you add a new resouce to your
configuration, the Akamai provider automatically determines the version of the configuration to which the
changes should be applied. If the latest version of the configuration is not active in either staging or
production, that version will be used. If the latest version is currently active, it will be cloned, and the
changes will be applied to the newly cloned version. (This version determination logic is used for all resources;
resource definitions in configuration files thus no longer include the version attribute that was required
under earlier versions of the provider.)

When you save the configuration file above and run `terraform apply`, Terraform updates the list of selected
hosts for the latest modifiable version of the specified configuration, and outputs the new list as values
for `selected_hostnames_appended`. The changes will be applied either to the most recent version (if it is not
active) or to a clone of that version.

After completing your changes to a security configuration, you can activate it in staging.

## Activate a Configuration

To activate the most recent version of a configuration, use the
[`akamai_appsec_activations`](../resources/appsec_activations.md) resource. 

Add the following resource block to your `akamai.tf` file:

```hcl
resource "akamai_appsec_activations" "activation" {
  config_id = data.akamai_appsec_configuration.configuration.config_id
  network = "STAGING"
  notes  = "TEST Notes"
  notification_emails = [ "my_name@mycompany.com" ]
}
```

After you save the file and run `terraform apply`, Terraform activates the configuration version in staging.
Upon completion of the activation, an email is sent to the addresses specified in the `notification_emails` list.

## Import Additional Resources

Terraform allows you to add a resource to its state even if this resource was created outside of Terraform,
for example by using the Control Center application. This allows you to keep Terraform's state in sync with
the state of your actual infrastructure. To do this, use the `terraform import` command with a configuration
file that includes a description of the existing resource. The `import` command requires that you specify both
the `address` and `ID` of the resource. The `address` indicates the destination to which the resource should
be imported; typically this is the resource type and local name of the resource as described in the local
configuration file. The `ID` indicates the unique identifier for this resource within Terraform's state. Its
format varies depending on the resource type, but it is generally formed by combining the values of the
resource's required parameters with a `:` separator, starting with the more general parameters. For example,
a security policy resource might have as an ID the string "33673:XYZ_12345", indicating security policy ID
`XYZ_12345` within configuration `33673`. You could use the information available in the Control Center to
create a matching description of this policy in your local configuration file. However, an easier way is to
use the `search` and `output_text` attributes of the `akamai_appsec_export_configuration` data source, which
will generate output for any of a list of resource types that can then be used as input on the `terraform
import` command line. The provider will automatically supply reasonable local names for the existing resources,
and will generate unique ID values using values from your configuration. For example, consider the following
configuration file:

```hcl
provider "akamai" {
  edgerc = "~/.edgerc"
}

// USE CASE: user wants to export a given configuration and version in tabular form 
data "akamai_appsec_configuration" "configuration" {
  name = var.security_configuration
}
data "akamai_appsec_export_configuration" "export" {
  config_id = data.akamai_appsec_configuration.configuration.config_id
  search = [
    "SelectedHostname.tf"
  ]
}
output "exported_configuration_text" {
  value = data.akamai_appsec_export_configuration.export.output_text
}
```

The `akamai_appsec_export_configuration` data source's optional `search` attribute will cause the Akamai
provider to examine your existing configuration and output descriptions of the existing resources of any
types listed (only one type is used in this example). Each resource definition is preceded by a comment
giving the command that can be run to import that resource instance. Running `terraform apply` with the
above command line would generate output like the following:

```hcl
// terraform import akamai_appsec_selected_hostnames.akamai_appsec_selected_hostname 12345
resource "akamai_appsec_selected_hostnames" "akamai_appsec_selected_hostname" {
 config_id = 12345
 mode = "REPLACE"
 hostnames = ["www.example.com","example.example.com"]
 }
```

Adding this output to a configuration file and running the `terraform import` command given in the comment
would then cause the `hostnames` list to be added to Terraform's local state.

Resource definitions can be exported for import using any of the following `search` entries:
  * AdvancedSettingsLogging.tf
  * AdvancedSettingsPragmaHeader.tf
  * AdvancedSettingsPrefetch.tf
  * ApiRequestConstraints.tf
  * CustomDeny.tf
  * CustomRule.tf
  * CustomRuleAction.tf
  * MatchTarget.tf
  * PenaltyBox.tf
  * RatePolicy.tf
  * RatePolicyAction.tf
  * ReputationProfile.tf
  * ReputationProfileAction.tf
  * Rule.tf
  * EvalRule.tf
  * EvalGroup.tf
  * AttackGroup.tf
  * SecurityPolicy.tf
  * SelectedHostname.tf
  * SiemSettings.tf
  * SlowPost.tf
  * IPGeoFirewall.tf
  * ThreatIntel.tf

## Data Sources & Resources Supported by the Akamai Appsec Provider

### Data Sources
  * akamai_appsec_advanced_settings_logging
  * akamai_appsec_advanced_settings_pragma_header
  * akamai_appsec_advanced_settings_prefetch
  * akamai_appsec_api_endpoints
  * akamai_appsec_api_request_constraints
  * akamai_appsec_attack_groups
  * akamai_appsec_bypass_network_lists
  * akamai_appsec_configuration
  * akamai_appsec_configuration_version
  * akamai_appsec_contracts_groups
  * akamai_appsec_custom_deny
  * akamai_appsec_custom_rule_actions
  * akamai_appsec_custom_rules
  * akamai_appsec_eval
  * akamai_appsec_eval_groups
  * akamai_appsec_eval_hostnames
  * akamai_appsec_eval_rules
  * akamai_appsec_export_configuration
  * akamai_appsec_failover_hostnames
  * akamai_appsec_hostname_coverage
  * akamai_appsec_hostname_coverage_match_targets
  * akamai_appsec_hostname_coverage_overlapping
  * akamai_appsec_ip_geo
  * akamai_appsec_match_targets
  * akamai_appsec_penalty_box
  * akamai_appsec_rate_policies
  * akamai_appsec_rate_policy_actions
  * akamai_appsec_reputation_profile_actions
  * akamai_appsec_reputation_profile_analysis
  * akamai_appsec_reputation_profiles
  * akamai_appsec_rule_upgrade_details
  * akamai_appsec_rules
  * akamai_appsec_security_policy
  * akamai_appsec_security_policy_protections
  * akamai_appsec_selectable_hostnames
  * akamai_appsec_selected_hostnames
  * akamai_appsec_siem_definitions
  * akamai_appsec_siem_settings
  * akamai_appsec_slow_post
  * akamai_appsec_version_notes
  * akamai_appsec_waf_mode
  * akamai_appsec_threat_intel

### Resources
  * akamai_appsec_activations
  * akamai_appsec_advanced_settings_logging
  * akamai_appsec_advanced_settings_pragma_header
  * akamai_appsec_advanced_settings_prefetch
  * akamai_appsec_api_constraints_protection
  * akamai_appsec_api_request_constraints
  * akamai_appsec_attack_group
  * akamai_appsec_bypass_network_lists
  * akamai_appsec_configuration
  * akamai_appsec_configuration_rename
  * akamai_appsec_custom_deny
  * akamai_appsec_custom_rule
  * akamai_appsec_custom_rule_action
  * akamai_appsec_eval
  * akamai_appsec_eval_group
  * akamai_appsec_eval_hostnames
  * akamai_appsec_eval_protect_host
  * akamai_appsec_eval_rule
  * akamai_appsec_ip_geo
  * akamai_appsec_ip_geo_protection
  * akamai_appsec_match_target
  * akamai_appsec_match_target_sequence
  * akamai_appsec_penalty_box
  * akamai_appsec_rate_policy
  * akamai_appsec_rate_policy_action
  * akamai_appsec_rate_protection
  * akamai_appsec_reputation_profile
  * akamai_appsec_reputation_profile_action
  * akamai_appsec_reputation_profile_analysis
  * akamai_appsec_reputation_protection
  * akamai_appsec_rule
  * akamai_appsec_rule_upgrade
  * akamai_appsec_security_policy
  * akamai_appsec_security_policy_rename
  * akamai_appsec_selected_hostnames
  * akamai_appsec_siem_settings
  * akamai_appsec_slow_post
  * akamai_appsec_slowpost_protection
  * akamai_appsec_version_notes
  * akamai_appsec_waf_mode
  * akamai_appsec_waf_protection
  * akamai_appsec_threat_intel

