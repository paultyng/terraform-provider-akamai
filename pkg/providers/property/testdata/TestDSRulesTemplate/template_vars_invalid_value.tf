provider "akamai" {
  edgerc = "~/.edgerc"
}

data "akamai_property_rules_template" "test" {
  template_file = "testdata/TestDSRulesTemplate/rules/property-snippets/template_in.json"
  variables {
    name = "criteriaMustSatisfy"
    value = "all"
    type = "number"
  }
  variables {
    name = "name"
    value = "test"
    type = "string"
  }
  variables {
    name = "enabled"
    value = "true"
    type = "bool"
  }
  variables {
    name = "list"
    value = "[\"foo\", \"bar\", \"baz\"]"
    type = "jsonArray"
  }
  variables {
    name = "options"
    value = "{\"enabled\":true}"
    type = "jsonBlock"
  }
}
