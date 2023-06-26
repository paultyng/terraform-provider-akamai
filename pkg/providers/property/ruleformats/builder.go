package ruleformats

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/papi"
	"github.com/akamai/terraform-provider-akamai/v4/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v4/pkg/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iancoleman/strcase"
)

// RulesBuilder orchestrates the construction of papi.Rules from the terraform schema.
type RulesBuilder struct {
	schemaReader  *RulesSchemaReader
	typeMappings  map[string]any
	nameMappings  map[string]string
	shouldFlatten func(string) bool
}

const defaultRule = "default"

// NewBuilder returns a new RulesBuilder that uses the provided schema.ResourceData to construct papi.Rules.
func NewBuilder(d *schema.ResourceData) *RulesBuilder {
	schemaReader := NewRulesSchemaReader(d)
	ruleFormat := schemaReader.GetRuleFormat()

	return &RulesBuilder{
		schemaReader:  schemaReader,
		shouldFlatten: ShouldFlattenFunc(ruleFormat),
		typeMappings:  TypeMappings(ruleFormat),
		nameMappings:  NameMappings(ruleFormat),
	}
}

// Build returns papi.Rules built from the terraform schema.
//
//nolint:gocyclo
func (r RulesBuilder) Build() (*papi.Rules, error) {
	name, err := r.ruleName()
	if err != nil {
		return nil, err
	}

	variables, err := r.ruleVariables()
	if err != nil {
		return nil, err
	}
	if name != defaultRule && len(variables) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrOnlyForDefault, "variable")
	}

	criteriaMustSatisfy, err := r.ruleCriteriaMustSatisfy()
	if name == defaultRule && err == nil {
		return nil, fmt.Errorf("%w: %s", ErrNotForDefault, "criteria_must_satisfy")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	isSecure, err := r.ruleIsSecure()
	if name != defaultRule && err == nil {
		return nil, fmt.Errorf("%w: %s", ErrOnlyForDefault, "is_secure")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	advancedOverride, err := r.ruleAdvancedOverride()
	if name != defaultRule && err == nil {
		return nil, fmt.Errorf("%w: %s", ErrOnlyForDefault, "advanced_override")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	comments, err := r.ruleComments()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	criteriaLocked, err := r.ruleCriteriaLocked()
	if name == defaultRule && err == nil {
		return nil, fmt.Errorf("%w: %s", ErrNotForDefault, "criteria_locked")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	customOverride, err := r.ruleCustomOverride()
	if name != defaultRule && err == nil {
		return nil, fmt.Errorf("%w: %s", ErrOnlyForDefault, "custom_override")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	uuid, err := r.ruleUUID()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	templateUUID, err := r.ruleTemplateUUID()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	templateLink, err := r.ruleTemplateLink()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	criteria, err := r.ruleCriteria()
	if name == defaultRule && err == nil {
		return nil, fmt.Errorf("%w: %s", ErrNotForDefault, "criterion")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	behaviors, err := r.ruleBehaviors()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	children, err := r.ruleChildren()
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return nil, err
	}

	rules := &papi.Rules{
		AdvancedOverride: advancedOverride,
		Behaviors:        behaviors,
		Children:         children,
		Comments:         comments,
		Criteria:         criteria,
		CriteriaLocked:   criteriaLocked,
		CustomOverride:   customOverride,
		Name:             name,
		Options: papi.RuleOptions{
			IsSecure: isSecure,
		},
		UUID:                uuid,
		TemplateUuid:        templateUUID,
		TemplateLink:        templateLink,
		Variables:           variables,
		CriteriaMustSatisfy: papi.RuleCriteriaMustSatisfy(criteriaMustSatisfy),
	}

	return rules, nil
}

func (r RulesBuilder) ruleVariables() ([]papi.RuleVariable, error) {
	variableList, err := r.schemaReader.GetVariablesList()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return []papi.RuleVariable{}, nil
		}
		return nil, err
	}

	variables := make([]papi.RuleVariable, 0, len(variableList))
	for _, variable := range variableList {
		variables = append(variables, papi.RuleVariable{
			Name:        variable["name"].(string),
			Description: tools.StringPtr(variable["description"].(string)),
			Value:       tools.StringPtr(variable["value"].(string)),
			Sensitive:   variable["sensitive"].(bool),
			Hidden:      variable["hidden"].(bool),
		})
	}

	return variables, nil
}

func (r RulesBuilder) ruleCriteriaMustSatisfy() (string, error) {
	key := r.schemaReader.criteriaMustSatisfyKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleAdvancedOverride() (string, error) {
	key := r.schemaReader.advancedOverrideKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleComments() (string, error) {
	key := r.schemaReader.commentsKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleCriteriaLocked() (bool, error) {
	key := r.schemaReader.criteriaLockedKey()
	return r.schemaReader.getBool(key)
}

func (r RulesBuilder) ruleCustomOverride() (*papi.RuleCustomOverride, error) {
	key := r.schemaReader.customOverrideKey()
	return r.schemaReader.getCustomOverride(key)
}

func (r RulesBuilder) ruleName() (string, error) {
	key := r.schemaReader.nameKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleIsSecure() (bool, error) {
	key := r.schemaReader.isSecureKey()
	return r.schemaReader.getBool(key)
}

func (r RulesBuilder) ruleUUID() (string, error) {
	key := r.schemaReader.uuidKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleTemplateUUID() (string, error) {
	key := r.schemaReader.templateUUIDKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleTemplateLink() (string, error) {
	key := r.schemaReader.templateLinkKey()
	return r.schemaReader.getString(key)
}

func (r RulesBuilder) ruleBehaviors() ([]papi.RuleBehavior, error) {
	behaviorsList, err := r.schemaReader.GetBehaviorsList()
	if err != nil {
		return nil, err
	}
	return r.buildRuleBehaviors(behaviorsList)
}

func (r RulesBuilder) ruleCriteria() ([]papi.RuleBehavior, error) {
	criteriaMap, err := r.schemaReader.GetCriteriaList()
	if err != nil {
		return nil, err
	}
	return r.buildRuleBehaviors(criteriaMap)
}

func (r RulesBuilder) buildRuleBehaviors(behaviorsList []RuleItem) ([]papi.RuleBehavior, error) {
	behaviors := make([]papi.RuleBehavior, 0)
	for _, item := range behaviorsList {
		itemName := strcase.ToLowerCamel(item.Name)
		if name, ok := r.nameMappings[itemName]; ok {
			itemName = name
		}

		b := papi.RuleBehavior{
			Name:         itemName,
			Locked:       getFromMapAndDeleteOrDefault(item.Item, "locked", false),
			UUID:         getFromMapAndDeleteOrDefault(item.Item, "uuid", ""),
			TemplateUuid: getFromMapAndDeleteOrDefault(item.Item, "template_uuid", ""),
		}

		b.Options = r.remapOptionValues(itemName, r.mapKeysToCamelCase(item.Item))

		behaviors = append(behaviors, b)
	}

	return behaviors, nil
}

// remapOptionValues ensures that options for behaviorName are in the format expected by the API.
// It either converts list to object by taking the first element or uses type mappings in specific cases
// e.g. if the API expects different types for the same attribute, depending on the value.
// If no action is required, value is rewritten without any mutations.
func (r RulesBuilder) remapOptionValues(behaviorName string, options papi.RuleOptionsMap) papi.RuleOptionsMap {
	newRom := make(papi.RuleOptionsMap)

	for optionName, v := range options {
		optKey := fmt.Sprintf("%s.%s", behaviorName, optionName)
		optValKey := fmt.Sprintf("%s.%v", optKey, v)
		if r.shouldFlatten(optKey) {
			slc, ok := v.([]any)
			if !ok {
				panic("unexpected type for: " + optKey)
			}
			if len(slc) > 1 {
				panic("expected object type has len()>1: " + optKey)
			}
			if len(slc) == 1 {
				newRom[optionName] = slc[0]
			}
		} else if mappedType, ok := r.typeMappings[optValKey]; ok {
			newRom[optionName] = mappedType
		} else {
			newRom[optionName] = v
		}

		if v, ok := newRom[optionName].(map[string]interface{}); ok {
			newRom[optionName] = r.remapOptionValues(optKey, v)
		}

	}

	return newRom
}

func (r RulesBuilder) ruleChildren() ([]papi.Rules, error) {
	childrenList, err := r.schemaReader.GetChildrenList()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return []papi.Rules{}, nil
		}
		return nil, err
	}

	children := make([]papi.Rules, 0, len(childrenList))
	for _, childJSON := range childrenList {
		var child papi.RulesUpdate
		err := json.Unmarshal([]byte(childJSON), &child)
		if err != nil {
			return nil, err
		}
		children = append(children, child.Rules)
	}

	return children, nil
}

func getFromMapAndDeleteOrDefault[T any](m map[string]any, key string, def T) T {
	res, ok := m[key]
	if !ok || res == nil {
		return def
	}

	delete(m, key)
	return res.(T)
}

func (r RulesBuilder) mapKeysToCamelCase(old map[string]any) map[string]any {
	newMap := make(map[string]any)
	for k, v := range old {
		if reflect.ValueOf(v).IsValid() {
			if mapValue, ok := v.(map[string]any); ok {
				v = r.mapKeysToCamelCase(mapValue)
			}
			if sliceValue, ok := v.([]any); ok {
				newSlice := make([]any, 0, len(sliceValue))
				for _, value := range sliceValue {
					if mapValue, ok := value.(map[string]any); ok {
						value = r.mapKeysToCamelCase(mapValue)
					}
					newSlice = append(newSlice, value)
				}
				v = newSlice
			}
			key := strcase.ToLowerCamel(k)
			if name, ok := r.nameMappings[key]; ok {
				key = name
			}
			newMap[key] = v
		}
	}
	return newMap
}
