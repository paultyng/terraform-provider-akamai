provider "akamai" {
  edgerc = "~/.edgerc"
}


data "akamai_siteshield_map" "test" {
   map_id = 1234
}