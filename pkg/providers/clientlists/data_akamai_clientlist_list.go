package clientlists

import (
	"context"
	"encoding/json"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/clientlists"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/hash"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceClientList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClientListRead,
		Schema: map[string]*schema.Schema{
			"list_id": {
				Type:        schema.TypeString,
				Required:    true,
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
			"items": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A set of client list values.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client list item creation date.",
						},
						"created_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the user who created the client list item.",
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
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The description of the client list item.",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client list item expiration date.",
						},
						"list_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the client list.",
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
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(getValidListTypes(), false)),
							},
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Value of the client list entry.",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A list of tags associated with the client list item.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
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
			"json": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON representation of the client list.",
			},
			"output_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tabular representation of the client lists.",
			},
		},
	}
}

func dataSourceClientListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	client := inst.Client(meta)
	logger := meta.Log("CLIENTLIST", "dataSourceClientListRead")

	listId, err := tf.GetStringValue("list_id", d)
	if err != nil {
		return diag.FromErr(err)
	}

	list, err := client.GetClientList(ctx, clientlists.GetClientListRequest{
		ListID:       listId,
		IncludeItems: true,
	})
	if err != nil {
		logger.Errorf("calling 'GetClientList': %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("list_id", list.ListID); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("version", list.Version); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("items_count", list.ItemsCount); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("create_date", list.CreateDate); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("created_by", list.CreatedBy); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("update_date", list.UpdateDate); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("updated_by", list.UpdatedBy); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("production_activation_status", list.ProductionActivationStatus); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("staging_activation_status", list.StagingActivationStatus); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("list_type", list.ListType); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("shared", list.Shared); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("read_only", list.ReadOnly); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}
	if err := d.Set("deprecated", list.Deprecated); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	mappedItems := mapClientListItemsToSchema(list)
	if err := d.Set("items", mappedItems); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	jsonBody, err := json.MarshalIndent(list.Items, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(jsonBody)); err != nil {
		return diag.Errorf("%v: %s", tf.ErrValueSet, err.Error())
	}

	d.SetId(hash.GetSHAString(string(jsonBody)))

	return nil
}

func mapClientListItemsToSchema(lists *clientlists.GetClientListResponse) []interface{} {
	if lists != nil && len(lists.Items) > 0 {
		result := make([]interface{}, 0, len(lists.Items))

		for _, list := range lists.Items {
			result = append(result, map[string]interface{}{
				"value":                        list.Value,
				"tags":                         list.Tags,
				"description":                  list.Description,
				"expiration_date":              list.ExpirationDate,
				"create_date":                  list.CreateDate,
				"created_by":                   list.CreatedBy,
				"production_activation_status": list.ProductionStatus,
				"staging_activation_status":    list.StagingStatus,
				"type":                         list.Type,
				"update_date":                  list.UpdateDate,
				"updated_by":                   list.UpdatedBy,
			})
		}

		return result
	}

	return make([]interface{}, 0)
}
