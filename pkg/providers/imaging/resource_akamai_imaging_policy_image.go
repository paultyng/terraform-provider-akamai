package imaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/imaging"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/session"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/logger"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// MaxPolicyDepth is the maximum supported depth of the nested transformations, before we reach hard limit of the Terraform
// provider for the gRPC communication for exchanging schema information, which is 64MB
const MaxPolicyDepth = 7

// PolicyDepth is variable to allow changing it for the unit tests, to achieve faster execution for tests,
// which do not need all supported levels
var PolicyDepth = MaxPolicyDepth

func resourceImagingPolicyImage() *schema.Resource {
	return &schema.Resource{
		CustomizeDiff: customdiff.All(
			enforcePolicyImageVersionChange,
		),
		CreateContext: resourcePolicyImageCreate,
		ReadContext:   resourcePolicyImageRead,
		UpdateContext: resourcePolicyImageUpdate,
		DeleteContext: resourcePolicyImageDelete,
		Schema: map[string]*schema.Schema{
			"activate_on_production": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "With this flag set to false, the user can perform modifications on staging without affecting the version already saved to production. " +
					"With this flag set to true, the policy will be saved on the production network. " +
					"It is possible to change it back to false only when there are any changes to the policy qualifying it for the new version.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The version number of this policy version",
			},
			"contract_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the Akamai Contract containing the Policy Set(s)",
			},
			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for a Policy. It is not possible to modify the id of the policy. Use .auto to modify the default policy.",
			},
			"policyset_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the Image & Video Manager Policy Set.",
			},
			"json": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsJSON),
				DiffSuppressFunc: diffSuppressPolicyImage,
				Description:      "A JSON encoded policy",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyImageImport,
		},
	}
}

func resourcePolicyImageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Imaging", "resourcePolicyImageCreate")
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)
	client := inst.Client(meta)
	logger.Debug("Creating policy")

	return upsertPolicyImage(ctx, d, m, client)
}

func upsertPolicyImage(ctx context.Context, d *schema.ResourceData, m interface{}, client imaging.Imaging) diag.Diagnostics {
	policyID, err := tf.GetStringValue("policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	contractID, err := tf.GetStringValue("contract_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	policySetID, err := tf.GetStringValue("policyset_id", d)
	if err != nil {
		return diag.FromErr(err)
	}

	var policy imaging.PolicyInputImage
	policyJSON, err := tf.GetStringValue("json", d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return diag.FromErr(err)
	}
	// upsert on staging only when there was a change
	if d.HasChangesExcept("activate_on_production") {
		upsertPolicyRequest := imaging.UpsertPolicyRequest{
			PolicyID:    policyID,
			Network:     imaging.PolicyNetworkStaging,
			ContractID:  contractID,
			PolicySetID: policySetID,
			PolicyInput: &policy,
		}
		createPolicyResp, err := client.UpsertPolicy(ctx, upsertPolicyRequest)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(fmt.Sprintf("%s:%s", policySetID, createPolicyResp.ID))
	}
	activateOnProduction, err := tf.GetBoolValue("activate_on_production", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return diag.FromErr(err)
	}
	if activateOnProduction {
		upsertPolicyRequest := imaging.UpsertPolicyRequest{
			PolicyID:    policyID,
			Network:     imaging.PolicyNetworkProduction,
			ContractID:  contractID,
			PolicySetID: policySetID,
			PolicyInput: &policy,
		}
		_, err := client.UpsertPolicy(ctx, upsertPolicyRequest)
		if err != nil {
			if errRestore := tf.RestoreOldValues(d, []string{"json"}); errRestore != nil {
				return diag.Errorf("%s\n%s: %s", err.Error(), "Failed to restore old state", errRestore.Error())
			}
			return diag.FromErr(err)
		}
	}

	return resourcePolicyImageRead(ctx, d, m)
}

func resourcePolicyImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Imaging", "resourcePolicyImageRead")
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)
	client := inst.Client(meta)

	logger.Debug("Reading policy")
	policyID, err := tf.GetStringValue("policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	contractID, err := tf.GetStringValue("contract_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	policysetID, err := tf.GetStringValue("policyset_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	network, err := getPolicyNetwork(d)
	if err != nil {
		return diag.FromErr(err)
	}
	policy, err := getPolicyImage(ctx, client, policyID, contractID, policysetID, network)
	if err != nil {
		return diag.FromErr(err)
	}
	policyInput, err := repackPolicyImageOutputToInput(policy)
	if err != nil {
		return diag.FromErr(err)
	}

	policyInput.RolloutDuration, err = getNotUpdateableField(d, extractRolloutDuration)
	if err != nil {
		return diag.FromErr(err)
	}

	extractServeStaleDuration := func(input imaging.PolicyInputImage) *int {
		return input.ServeStaleDuration
	}

	policyInput.ServeStaleDuration, err = getNotUpdateableField(d, extractServeStaleDuration)
	if err != nil {
		return diag.FromErr(err)
	}

	policyJSON, err := getPolicyImageJSON(policyInput)
	if err != nil {
		return diag.FromErr(err)
	}

	attrs := make(map[string]interface{})
	attrs["json"] = policyJSON
	attrs["version"] = policy.Version
	if err := tf.SetAttrs(d, attrs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getPolicyNetwork(d *schema.ResourceData) (networkType imaging.PolicyNetwork, err error) {
	network := imaging.PolicyNetworkStaging
	activateOnProduction, err := tf.GetBoolValue("activate_on_production", d)
	if err != nil {
		return "", err
	}
	if activateOnProduction {
		network = imaging.PolicyNetworkProduction
	}
	return network, nil
}

var extractRolloutDuration = func(input imaging.PolicyInputImage) *int {
	return input.RolloutDuration
}

func repackPolicyImageOutputToInput(policy *imaging.PolicyOutputImage) (*imaging.PolicyInputImage, error) {
	policyOutputJSON, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	policyInput := imaging.PolicyInputImage{}
	err = json.Unmarshal(policyOutputJSON, &policyInput)
	if err != nil {
		return nil, err
	}
	return &policyInput, nil
}

func getPolicyImage(ctx context.Context, client imaging.Imaging, policyID, contractID, policySetID string, network imaging.PolicyNetwork) (*imaging.PolicyOutputImage, error) {
	policyRequest := imaging.GetPolicyRequest{
		PolicyID:    policyID,
		Network:     network,
		ContractID:  contractID,
		PolicySetID: policySetID,
	}
	policyOutput, err := client.GetPolicy(ctx, policyRequest)
	if err != nil {
		return nil, err
	}
	policy, ok := policyOutput.(*imaging.PolicyOutputImage)
	if !ok {
		return nil, fmt.Errorf("policy is not of type image")
	}

	return policy, nil
}

func getPolicyImageJSON(policy *imaging.PolicyInputImage) (string, error) {
	policyJSON, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return "", err
	}
	return string(policyJSON), nil
}

func resourcePolicyImageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Imaging", "resourcePolicyImageUpdate")
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)
	client := inst.Client(meta)
	logger.Debug("Updating policy")

	return upsertPolicyImage(ctx, d, m, client)
}

func resourcePolicyImageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Imaging", "resourcePolicyImageDelete")
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)
	client := inst.Client(meta)
	logger.Debug("Deleting policy")

	policyID, err := tf.GetStringValue("policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}

	// .auto policy cannot be removed alone, only via removal of policy set
	if policyID == ".auto" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Image and Video Manager API does not support '.auto' policy deletion, it can be removed when removing related policy set - resource will only be removed from state.",
			},
		}
	}

	contractID, err := tf.GetStringValue("contract_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	policySetID, err := tf.GetStringValue("policyset_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	// delete on staging
	deletePolicyRequest := imaging.DeletePolicyRequest{
		PolicyID:    policyID,
		Network:     imaging.PolicyNetworkStaging,
		ContractID:  contractID,
		PolicySetID: policySetID,
	}
	_, err = client.DeletePolicy(ctx, deletePolicyRequest)
	if err != nil {
		return diag.FromErr(err)
	}
	// delete on production, API returns success even if policy does not exist
	// it is faster to attempt to delete on production than checking if there is policy on production first
	deletePolicyRequest = imaging.DeletePolicyRequest{
		PolicyID:    policyID,
		Network:     imaging.PolicyNetworkProduction,
		ContractID:  contractID,
		PolicySetID: policySetID,
	}
	_, err = client.DeletePolicy(ctx, deletePolicyRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourcePolicyImageImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	meta := meta.Must(m)
	logger := meta.Log("Imaging", "resourcePolicyImageImport")
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)
	client := inst.Client(meta)

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("colon-separated list of policy ID, policy set ID and contract ID has to be supplied in import: %s", d.Id())
	}
	policyID, policySetID, contractID := parts[0], parts[1], parts[2]

	policyStaging, err := getPolicyImage(ctx, client, policyID, contractID, policySetID, imaging.PolicyNetworkStaging)
	if err != nil {
		return nil, err
	}
	policy, err := repackPolicyImageOutputToInput(policyStaging)
	if err != nil {
		return nil, err
	}
	policyStagingJSON, err := getPolicyImageJSON(policy)
	if err != nil {
		return nil, err
	}

	var activateOnProduction bool
	policyProduction, err := getPolicyImage(ctx, client, policyID, contractID, policySetID, imaging.PolicyNetworkProduction)
	if err != nil {
		var e *imaging.Error
		if ok := errors.As(err, &e); !ok || e.Status != http.StatusNotFound {
			return nil, err
		}
	} else {
		policy, err := repackPolicyImageOutputToInput(policyProduction)
		if err != nil {
			return nil, err
		}
		policyProductionJSON, err := getPolicyImageJSON(policy)
		if err != nil {
			return nil, err
		}
		activateOnProduction = equalPolicyImage(policyStagingJSON, policyProductionJSON)
	}

	attrs := make(map[string]interface{})
	attrs["policy_id"] = policyID
	attrs["contract_id"] = contractID
	attrs["policyset_id"] = policySetID
	attrs["activate_on_production"] = activateOnProduction
	if err := tf.SetAttrs(d, attrs); err != nil {
		return nil, err
	}

	d.SetId(fmt.Sprintf("%s:%s", policySetID, policyID))

	return []*schema.ResourceData{d}, nil
}

func diffSuppressPolicyImage(_, old, new string, _ *schema.ResourceData) bool {
	return equalPolicyImage(old, new)
}

func equalPolicyImage(old, new string) bool {
	logger := logger.Get("Imaging", "equalPolicyImage")
	if old == new {
		return true
	}
	var oldPolicy, newPolicy imaging.PolicyInputImage
	if old == "" || new == "" {
		return old == new
	}
	if err := json.Unmarshal([]byte(old), &oldPolicy); err != nil {
		logger.Errorf("Unable to unmarshal 'old' JSON policy: %s", err)
		return false
	}
	if err := json.Unmarshal([]byte(new), &newPolicy); err != nil {
		logger.Errorf("Unable to unmarshal 'new' JSON policy: %s", err)
		return false
	}

	sortPolicyImageFields(&oldPolicy, &newPolicy)

	return reflect.DeepEqual(oldPolicy, newPolicy)
}

// enforcePolicyImageVersionChange enforces that change to any field will most likely result in creating a new version
func enforcePolicyImageVersionChange(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	o, n := diff.GetChange("json")

	oldValue := o.(string)
	newValue := n.(string)

	if diff.HasChange("contract_id") ||
		diff.HasChange("policy_id") ||
		diff.HasChange("policyset_id") ||
		!equalPolicyImage(oldValue, newValue) {
		return diff.SetNewComputed("version")
	}
	return nil
}

func getNotUpdateableField(d *schema.ResourceData, extractionFunc func(input imaging.PolicyInputImage) *int) (*int, error) {
	inputJSON, err := tf.GetStringValue("json", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		input := imaging.PolicyInputImage{}
		err = json.Unmarshal([]byte(inputJSON), &input)
		if err != nil {
			return nil, err
		}
		return extractionFunc(input), nil
	}
	return nil, nil
}

// sortPolicyImageFields sorts any fields that are present in image policy and might cause diffs
func sortPolicyImageFields(oldPolicy, newPolicy *imaging.PolicyInputImage) {
	sortPolicyImageOutput(oldPolicy)
	sortPolicyImageOutput(newPolicy)

	sortPolicyHosts(oldPolicy)
	sortPolicyHosts(newPolicy)

	sortPolicyVariables(oldPolicy)
	sortPolicyVariables(newPolicy)

	sortPolicyBreakpointsWidths(oldPolicy)
	sortPolicyBreakpointsWidths(newPolicy)
}

// sortPolicyImageOutput sorts PolicyInputImage's output allowedFormats and forcedFormats
func sortPolicyImageOutput(policy *imaging.PolicyInputImage) {
	if policy.Output != nil {
		if policy.Output.AllowedFormats != nil {
			allowedFormats := policy.Output.AllowedFormats
			sort.Slice(allowedFormats, func(i, j int) bool {
				return allowedFormats[i] < allowedFormats[j]
			})
			policy.Output.AllowedFormats = allowedFormats
		}

		if policy.Output.ForcedFormats != nil {
			forcedFormats := policy.Output.ForcedFormats
			sort.Slice(forcedFormats, func(i, j int) bool {
				return forcedFormats[i] < forcedFormats[j]
			})
			policy.Output.ForcedFormats = forcedFormats
		}
	}
}

// sortPolicyHosts sorts PolicyInputImage's hosts
func sortPolicyHosts(policy *imaging.PolicyInputImage) {
	if policy.Hosts != nil {
		sort.Strings(policy.Hosts)
	}
}

// sortPolicyVariables sorts PolicyInputImage's variables
func sortPolicyVariables(policy *imaging.PolicyInputImage) {
	if policy.Variables != nil {
		variables := policy.Variables
		sort.Slice(variables, func(i, j int) bool {
			return variables[i].Name < variables[j].Name
		})
		policy.Variables = variables
	}
}

// sortPolicyBreakpointsWidths sorts PolicyInputImage's breakpoints.Widths
func sortPolicyBreakpointsWidths(policy *imaging.PolicyInputImage) {
	if policy.Breakpoints != nil && policy.Breakpoints.Widths != nil {
		sort.Ints(policy.Breakpoints.Widths)
	}
}
