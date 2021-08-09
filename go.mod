module github.com/akamai/terraform-provider-akamai/v2

require (
	github.com/akamai/AkamaiOPEN-edgegrid-golang/v2 v2.5.0
	github.com/allegro/bigcache v1.2.1
	github.com/apex/log v1.9.0
	github.com/aws/aws-sdk-go v1.31.9 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.1.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-hclog v0.15.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/jedib0t/go-pretty/v6 v6.0.4
	github.com/jinzhu/copier v0.3.0
	github.com/spf13/cast v1.3.1
	github.com/stretchr/testify v1.7.0
	github.com/tj/assert v0.0.3
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/tools v0.1.3 // indirect
	google.golang.org/api v0.34.0 // indirect
)

replace (
	//github.com/akamai/AkamaiOPEN-edgegrid-golang/v2 => ../AkamaiOPEN-edgegrid-golang

	// https://github.com/golang/lint/issues/446
	github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	// https://github.com/sourcegraph/go-diff/issues/33
	github.com/sourcegraph/go-diff => github.com/sourcegraph/go-diff v0.5.1
	sourcegraph.com/sourcegraph/go-diff => github.com/sourcegraph/go-diff v0.5.1
)

go 1.14
