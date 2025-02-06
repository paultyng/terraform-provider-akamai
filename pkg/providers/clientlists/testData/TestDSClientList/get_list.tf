provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_clientlist_list" "list" {
  list_id = "123_TEST"
}
