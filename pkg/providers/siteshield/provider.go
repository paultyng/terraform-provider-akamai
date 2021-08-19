package siteshield

import (
	"sync"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/siteshield"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/akamai"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/config"
	"github.com/apex/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type (
	provider struct {
		*schema.Provider

		client siteshield.SSMAPS
	}
	// Option is a siteshield provider option
	Option func(p *provider)
)

var (
	once sync.Once

	inst *provider
)

// Subprovider returns a core sub provider
func Subprovider(opts ...Option) akamai.Subprovider {
	once.Do(func() {
		inst = &provider{Provider: Provider()}

		for _, opt := range opts {
			opt(inst)
		}
	})

	return inst
}

// Provider returns the Akamai terraform.Resource provider.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"siteshield": {
				Optional: true,
				Type:     schema.TypeSet,
				Elem:     config.Options("siteshield"),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"akamai_siteshield_map": dataSourceSiteShieldMap(),
		},
	}
	return provider
}

// WithClient sets the client interface function, used for mocking and testing
func WithClient(c siteshield.SSMAPS) Option {
	return func(p *provider) {
		p.client = c
	}
}

// Client returns the PAPI interface
func (p *provider) Client(meta akamai.OperationMeta) siteshield.SSMAPS {
	if p.client != nil {
		return p.client
	}
	return siteshield.Client(meta.Session())
}

func getSiteShieldV1Service(d *schema.ResourceData) error {
	var section string

	if section != "" {
		if err := d.Set("config_section", section); err != nil {
			return err
		}
	}

	return nil
}

func (p *provider) Name() string {
	return "siteshield"
}

// SiteShieldProviderVersion update version string anytime provider adds new features
const SiteShieldProviderVersion string = "v0.0.1"

func (p *provider) Version() string {
	return SiteShieldProviderVersion
}

func (p *provider) Schema() map[string]*schema.Schema {
	return p.Provider.Schema
}

func (p *provider) Resources() map[string]*schema.Resource {
	return p.Provider.ResourcesMap
}

func (p *provider) DataSources() map[string]*schema.Resource {
	return p.Provider.DataSourcesMap
}

func (p *provider) Configure(log log.Interface, d *schema.ResourceData) diag.Diagnostics {
	log.Debug("START Configure")

	err := getSiteShieldV1Service(d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
