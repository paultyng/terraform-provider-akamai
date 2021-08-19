package siteshield

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/siteshield"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/akamai"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSiteShieldMap() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSiteShieldMapRead,
		Schema: map[string]*schema.Schema{
			"map_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"current_cidrs": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"proposed_cidrs": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"rule_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"acknowledged": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceSiteShieldMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("SSMAP", "dataSiteShieldMap")

	mapID, err := tools.GetIntValue("map_id", d)
	d.SetId(strconv.Itoa(mapID))

	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}

	ssMapID := siteshield.SiteShieldMapRequest{UniqueID: mapID}

	ssMap, err := client.GetSiteShieldMap(ctx, ssMapID)
	if err != nil {
		logger.Errorf("calling 'getSiteShieldMap': %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("current_cidrs", ssMap.CurrentCidrs); err != nil {
		logger.Errorf("error setting 'current_cidrs': %s", err.Error())
		return diag.FromErr(fmt.Errorf("%w: %s", tools.ErrValueSet, err.Error()))
	}

	if err := d.Set("proposed_cidrs", ssMap.ProposedCidrs); err != nil {
		logger.Errorf("error setting 'proposed_cidrs': %s", err.Error())
		return diag.FromErr(fmt.Errorf("%w: %s", tools.ErrValueSet, err.Error()))
	}

	if err := d.Set("rule_name", ssMap.RuleName); err != nil {
		logger.Errorf("error setting 'rule_name': %s", err.Error())
		return diag.FromErr(fmt.Errorf("%w: %s", tools.ErrValueSet, err.Error()))
	}

	if err := d.Set("acknowledged", ssMap.Acknowledged); err != nil {
		logger.Errorf("error setting 'acknowledged': %s", err.Error())
		return diag.FromErr(fmt.Errorf("%w: %s", tools.ErrValueSet, err.Error()))
	}

	return nil
}
