provider "akamai" {
  edgerc = "~/.edgerc"
}


data "akamai_group" "group" {
 group_name = "SOME GROUP NAME"
 contract_id = "SOME CONTRACT ID - KEEP THE ctr_ PREFIX"
}


data "akamai_contract" "contract" {
  group_name = data.akamai_group.group.name
}


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
