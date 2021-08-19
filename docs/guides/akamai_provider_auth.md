---
layout: "akamai"
page_title: "Akamai Provider: Set Up Authentication"
description: |-
  Learn how to set up authentication for the Akamai Terraform Provider.
---

# Authenticate the Akamai Provider

The Akamai Terraform Provider relies on Akamai's EdgeGrid authentication scheme. The Akamai Provider code acts as a
wrapper for Akamai's APIs and reuses the same authentication mechanism. Akamai recommends storing your API credentials in a local `.edgerc` file.

The permissions you need for the Akamai Provider depend on
the subset of Akamai resources and data sources you'll use. Without
these permissions, your Terraform configurations won't execute.

See [Get Started with APIs](https://developer.akamai.com/api/getting-started)
for more information on the process of creating Akamai credentials.


## Set up your API clients

To prepare the `.edgerc` file, you need to:

1.  Determine which Akamai Provider modules you want to use with Terraform.
2. Find the API service name for the modules you'll be adding and [create an API client](https://developer.akamai.com/api/getting-started#createanapiclient).
3.  If you already have the API clients you need, you can add the credentials to your local `.edgerc` file.

For example, if you want to add Akamai properties to your existing
Terraform configuration, you'll be using the Property Provisioning module. For this module,
 you need to create an API client for the **Property Manager (PAPI)** service,
 then add it to your `.edgerc` file.

Here's a list of the Akamai modules available on Terraform and their
supporting API service names:

| **Module** | **API service name** |
|-------------|----------------------|
| **Application Security** | Application Security |
| **Certificate Provisioning** | Certificate Provisioning System |
| **DNS Zone Administration Module** | DNS Zone Management |
| **Global Traffic Management Domain Administration** | Global Traffic Management |
| **Identity and Access Management** | Identity Management: User Administration |
| **Network Lists** | Network Lists |
| **Property Provisioning** (Includes Common functions) | Property Manager (PAPI) |
| **Site Shield Maps** | Site Shield Maps |

-> **Note:** If you're using the Edge DNS or GTM module, you may also need the Property Manager API service. Whether you need this additional service depends on your contract and group. See [PAPI concepts](https://developer.akamai.com/api/core_features/property_manager/v1.html#papiconcepts) for more information.

## Default authentication settings

You can start using the Akamai Provider without specifying any additional authentication details in the configuration file. By default, the provider looks for the credentials file in the `$HOME/.edgerc` directory and uses the `default` section to retrieve client tokens.

## Authenticate using custom credentials

If you don't want to use default settings, you can explicitly reference your local `.edgerc` file in the `akamai.tf` configuration file.

### Example usage

Terraform 0.13 and later:

```hcl
terraform {
  required_providers {
    akamai = {
      source  = "hashicorp/akamai"
    }
  }
}

# Configure the Akamai Provider
provider "akamai" {
  edgerc = "~/.edgerc"
  config_section = "custom"
}

# Create a property
resource "akamai_property" "example_property" {
  name = "www.example.org"

  # ...
}

# Create a DNS record
resource "akamai_dns_record" "example_record" {
  zone       = "example.org"
  name       = "www.example.org"
  recordtype = "CNAME"
  active     = true
  ttl        = 600
  target     = ["example.org.akamaized.net."]
}
```

Terraform 0.12 and earlier:

```hcl
# Configure the Akamai Provider
provider "akamai" {
  edgerc = "~/.edgerc"
  config_section = "custom"
}

# Create a property
resource "akamai_property" "example_property" {
  name = "www.example.org"

  # ...
}

# Create a DNS record
resource "akamai_dns_record" "example_record" {
  zone       = "example.org"
  name       = "www.example.org"
  recordtype = "CNAME"
  active     = true
  ttl        = 600
  target     = ["example.org.akamaized.net."]
}
```


### Argument reference

Arguments supported in the `provider` block:

* `edgerc` - (Optional) The location of the `.edgerc` file containing credentials. The default is `$HOME/.edgerc`.
* `config_section` - (Optional) The credential section to use within the `.edgerc` file for all EdgeGrid calls. If you don't specify the `config_section` argument, the Akamai Provider uses the credentials from the `default` section of the `.edgerc` file.

#### Deprecated arguments

* `property_section` - (Deprecated) The credential section to use for the [Property Manager API](https://developer.akamai.com/api/core_features/property_manager/v1.html).
If you don't use `property_section`, the Akamai Provider uses the credentials in the `default` section of the `.edgerc` file.
* `dns_section` - (Deprecated) The credential section to use for the [Edge DNS Zone Management API](https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html). If you don't use `dns_section`, the Akamai Provider uses the credentials in the `default` section of the `.edgerc` file.
* `gtm_section` - (Deprecated) The credential section to use for the [Global Traffic Management API](https://developer.akamai.com/api/web_performance/global_traffic_management/v1.html). If you don't use `gtm_section`, the Akamai Provider uses the credentials in the `default` section of the `.edgerc` file.

## Authenticate using inline credentials

You should generally use default settings or reference a local `.edgerc` file in the `akamai.tf` configuration to authenticate the Terraform Provider. However, if needed, you can specify inline credentials for each
resource or data source.

Under `provider`, add a `config` block that includes authentication details. You then don't need the `edgerc` or `config_section` attributes that you'd use if you were adding your local `.edgerc` file to your Akamai Provider configuration.

### Example usage

```
provider "akamai" {
  config { # NOTE : Replace values with valid edge token values
    client_secret = aaaaaaaaaaaaaaaaaaaa12345xyz=
    host = akaa-XXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXX.luna.akamaiapis.net
    access_token = akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx
    client_token = akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx
  }
}
```

### Argument reference

* `config` - (Optional) Provide credentials for the Akamai Provider. The block supports these arguments:
  * `host` - (Required) The base credential hostname without the protocol.
  * `access_token` - (Required) The service's access token from the `.edgerc` file.
  * `client_token` - (Required) The service's client token from the `.edgerc` file.
  * `client_secret` - (Required) The service's client secret from the `.edgerc` file.
  * `max_body` - (Optional) The service's maximum data payload size in bytes.
  * `account_key` - (Optional) If managing multiple accounts, the account ID you want to use when running Terraform commands. The account selected persists for all commands until you change it.

#### Deprecated arguments

* `dns` - (Deprecated) Legacy Edge DNS API service argument for inline authentication. Used same arguments as the current `config` block.
* `gtm` - (Deprecated) Legacy Global Traffic Management API service argument for inline authentication. Used same arguments as the current `config` block.
* `property` - (Deprecated) Legacy Property Manager API service argument for inline authentication. Used same arguments as the current `config` block.

## Authenticate using environment variables

You can also use environment variables to set credential values.
Environment variables take precedence over the settings in the `.edgerc`
file.

Use this template to specify variables based on the sections of your `.edgerc` file they represent: `AKAMAI{_SECTION_NAME}_*` .

### Example usage

To set up the Property Provisioning module, you'll need an API client for Property Manager.

1. Set up the basic `provider` block so that it's empty:

    ```
    provider "akamai" {}
    ```

2. In your terminal, specify the variable values for the credentials section you want to use and run `terraform apply`. This example includes variables for the `papi` credentials section:

    ```
    AKAMAI_PAPI_HOST=akaa-XXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXX.luna.akamaiapis.net \
    AKAMAI_PAPI_ACCESS_TOKEN=akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx \
    AKAMAI_PAPI_CLIENT_TOKEN=akaa-xxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxx \
    AKAMAI_PAPI_CLIENT_SECRET=aaaaaaaaaaaaaaaaaaaa12345xyz= \
    terraform apply
    ```

3. In the `akamai.tf` file, under `provider`, specify the `config_section` you set the variables for:

    ```
    provider "akamai" {
      config_section = "papi"
    ```

### Variable reference

Variable names correspond to the sections of your `.edgerc` file they represent. The environment variables should be in this format: `AKAMAI{_SECTION_NAME}_*`

These are the variables for the `default` section of your `.edgerc` and what they represent:

* `AKAMAI_HOST` - (Required) The base credential hostname without the protocol.
* `AKAMAI_ACCESS_TOKEN` - (Required) The service's access token from the `.edgerc` file.
* `AKAMAI_CLIENT_TOKEN` - (Required) The service's client token from the `.edgerc` file.
* `AKAMAI_CLIENT_SECRET` - (Required) The service's client secret from the `.edgerc` file.
* `AKAMAI_MAX_BODY` - (Optional) The service's maximum data payload size in bytes.
* `AKAMAI_ACCOUNT_KEY` - (Optional) If managing multiple accounts, the account ID you want to use when running Terraform commands. The account selected persists for all commands until you change it.
