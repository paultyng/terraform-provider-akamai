package clientlists

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v10/pkg/clientlists"
	"github.com/akamai/terraform-provider-akamai/v7/pkg/common/hash"
	"github.com/akamai/terraform-provider-akamai/v7/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v7/pkg/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceClientLists() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClientListsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(getValidListTypes(), false)),
				},
			},
			"list_ids": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A set of client list ids.",
			},
			"lists": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A set of client lists.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the client list",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The type of the client list",
						},
						"notes": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client list notes",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The client list tags",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"list_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the client list.",
						},
						"version": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The current version of the client list.",
						},
						"items_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of items that a client list contains.",
						},
						"create_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client list creation date.",
						},
						"created_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the user who created the client list.",
						},
						"update_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date of last update.",
						},
						"updated_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the user that updated the client list last.",
						},
						"production_activation_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The activation status in production environment.",
						},
						"staging_activation_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The activation status in staging environment.",
						},
						"list_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client list type.",
						},
						"shared": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the client list is shared.",
						},
						"read_only": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the client is editable for the authenticated user.",
						},
						"deprecated": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the client list was removed.",
						},
					},
				},
			},
			"json": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON representation of the client lists.",
			},
			"output_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tabular representation of the client lists.",
			},
		},
	}
}

func dataSourceClientListsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	client := inst.Client(meta)
	logger := meta.Log("CLIENTLIST", "dataSourceClientListRead")

	name, err := tf.GetStringValue("name", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return diag.FromErr(err)
	}

	listTypesSet, err := tf.GetSetValue("type", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return diag.FromErr(err)
	}
	listTypesList := tf.SetToStringSlice(listTypesSet)
	listTypes := make([]clientlists.ClientListType, 0, listTypesSet.Len())
	for _, v := range listTypesList {
		listTypes = append(listTypes, clientlists.ClientListType(v))
	}

	lists, err := client.GetClientLists(ctx, clientlists.GetClientListsRequest{
		Name: name,
		Type: listTypes,
	})
	if err != nil {
		logger.Errorf("calling 'GetClientLists': %s", err.Error())
		return diag.FromErr(err)
	}

	mappedLists := mapClientListsToSchema(lists)
	if err := d.Set("lists", mappedLists); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	jsonBody, err := json.MarshalIndent(lists.Content, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(jsonBody)); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	IDs := make([]string, 0, len(lists.Content))
	for _, cl := range lists.Content {
		IDs = append(IDs, cl.ListID)
	}
	if err := d.Set("list_ids", IDs); err != nil {
		logger.Errorf("error setting 'list_ids': %s", err.Error())
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	ots := OutputTemplates{}
	InitTemplates(ots)

	outputText, err := RenderTemplates(ots, "clientListsDS", lists)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("output_text", outputText); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	d.SetId(hash.GetSHAString(string(jsonBody)))

	return nil
}

func mapClientListsToSchema(lists *clientlists.GetClientListsResponse) []interface{} {
	if lists != nil && len(lists.Content) > 0 {
		result := make([]interface{}, 0, len(lists.Content))

		for _, list := range lists.Content {
			result = append(result, map[string]interface{}{
				"name":                         list.Name,
				"type":                         list.Type,
				"notes":                        list.Notes,
				"tags":                         list.Tags,
				"list_id":                      list.ListID,
				"version":                      list.Version,
				"items_count":                  list.ItemsCount,
				"create_date":                  list.CreateDate,
				"created_by":                   list.CreatedBy,
				"update_date":                  list.UpdateDate,
				"updated_by":                   list.UpdatedBy,
				"production_activation_status": list.ProductionActivationStatus,
				"staging_activation_status":    list.StagingActivationStatus,
				"list_type":                    list.ListType,
				"shared":                       list.Shared,
				"read_only":                    list.ReadOnly,
				"deprecated":                   list.Deprecated,
			})
		}

		return result
	}

	return make([]interface{}, 0)
}

func getValidListTypes() []string {
	return []string{
		string(clientlists.IP),
		string(clientlists.GEO),
		string(clientlists.ASN),
		string(clientlists.TLSFingerprint),
		string(clientlists.FileHash),
	}
}
