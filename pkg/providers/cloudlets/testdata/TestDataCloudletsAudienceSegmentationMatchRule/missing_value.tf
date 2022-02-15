provider "akamai" {
  edgerc = "~/.edgerc"
}

data "akamai_cloudlets_audience_segmentation_match_rule" "test" {
  match_rules {
    matches {
      match_type     = "clientip"
      match_operator = "equals"
    }
    forward_settings {
    }
  }
}