// Package clientlists contains implementation for Akamai Terraform sub-provider responsible for creation, deployment, and management of client lists
package clientlists

import (
	"sync"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v10/pkg/clientlists"
	"github.com/akamai/terraform-provider-akamai/v7/pkg/meta"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type (
	// Subprovider gathers clientlists resources and data sources
	Subprovider struct {
		client clientlists.ClientLists
	}
	// Option is a clientlists provider option
	Option func(p *Subprovider)
)

var (
	once sync.Once

	inst *Subprovider
)

// NewSubprovider returns a new clientlists subprovider
func NewSubprovider(opts ...Option) *Subprovider {
	once.Do(func() {
		inst = &Subprovider{}

		for _, opt := range opts {
			opt(inst)
		}
	})

	return inst
}

// WithClient sets the client interface function, used for mocking and testing
func WithClient(c clientlists.ClientLists) Option {
	return func(p *Subprovider) {
		p.client = c
	}
}

// Client returns the ClientLists interface
func (p *Subprovider) Client(meta meta.Meta) clientlists.ClientLists {
	if p.client != nil {
		return p.client
	}
	return clientlists.Client(meta.Session())
}

// SDKResources returns the clientlists resources implemented using terraform-plugin-sdk
func (p *Subprovider) SDKResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"akamai_clientlist_activation": resourceClientListActivation(),
		"akamai_clientlist_list":       resourceClientList(),
	}
}

// SDKDataSources returns the clientlists data sources implemented using terraform-plugin-sdk
func (p *Subprovider) SDKDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"akamai_clientlist_lists": dataSourceClientLists(),
		"akamai_clientlist_list":  dataSourceClientList(),
	}
}

// FrameworkResources returns the clientlists resources implemented using terraform-plugin-framework
func (p *Subprovider) FrameworkResources() []func() resource.Resource {
	return []func() resource.Resource{}
}

// FrameworkDataSources returns the clientlists data sources implemented using terraform-plugin-framework
func (p *Subprovider) FrameworkDataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
