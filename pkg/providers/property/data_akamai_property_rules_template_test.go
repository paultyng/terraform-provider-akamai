package property

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/akamai/terraform-provider-akamai/v2/pkg/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func TestDataAkamaiPropertyRulesRead(t *testing.T) {
	t.Run("valid nested template with vars map", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSRulesTemplate/template_vars_map.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_property_rules_template.test", "json", loadFixtureString("testdata/TestDSRulesTemplate/rules/rules_out.json")),
						),
					},
				},
			})
		})
	})
	t.Run("valid nested template with vars files", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSRulesTemplate/template_null_values.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_property_rules_template.test", "json", loadFixtureString("testdata/TestDSRulesTemplate/rules/rules_defaults.json")),
						),
					},
				},
			})
		})
	})
	t.Run("null values do not overwrite defaults", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSRulesTemplate/template_vars_file.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_property_rules_template.test", "json", loadFixtureString("testdata/TestDSRulesTemplate/rules/rules_out.json")),
						),
					},
				},
			})
		})
	})
	t.Run("error setting both ,ap and file variables", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_vars_conflict.tf"),
						ExpectError: regexp.MustCompile(`"variables": conflicts with var_definition_file`),
					},
				},
			})
		})
	})
	t.Run("variable has invalid type", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_vars_invalid_type.tf"),
						ExpectError: regexp.MustCompile(`'type' has invalid value: should be 'bool', 'number', 'string', 'jsonArray' or 'jsonBlock'`),
					},
				},
			})
		})
	})
	t.Run("variable not found in template", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_var_not_found.tf"),
						ExpectError: regexp.MustCompile(`executing "snippets/sub/another-template.json" at <.options>: map has no entry for key "options"`),
					},
				},
			})
		})
	})
	t.Run("invalid variable in map", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_vars_invalid_value.tf"),
						ExpectError: regexp.MustCompile(`value could not be represented as number: strconv.ParseFloat: parsing "all": invalid syntax`),
					},
				},
			})
		})
	})
	t.Run("error fetching vars from map - file not found", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_vars_file_not_found.tf"),
						ExpectError: regexp.MustCompile(`reading file: open invalid_path: no such file or directory`),
					},
				},
			})
		})
	})
	t.Run("invalid json result", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_invalid_json.tf"),
						ExpectError: regexp.MustCompile(`invalid JSON result:`),
					},
				},
			})
		})
	})
	t.Run("template file not found", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_file_not_found.tf"),
						ExpectError: regexp.MustCompile(`Error: stat testdata/TestDSRulesTemplate/rules/property-snippets/non-existent.json: no such file or directory`),
					},
				},
			})
		})
	})

	t.Run("snippets files are under incorrect folder deeply nested", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_invalid_snippets_folder_json.tf"),
						ExpectError: regexp.MustCompile(`Error: snippets file should be under 'property-snippets' folder with .json extension. Invalid file: testdata/TestDSRulesTemplate/output/template_invalid_json.json`),
					},
				},
			})
		})
	})

	t.Run("snippets files are under incorrect folder e.g. property-snippets/rules.json", func(t *testing.T) {
		client := mockpapi{}
		useClient(&client, func() {
			resource.UnitTest(t, resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      loadFixtureString("testdata/TestDSRulesTemplate/template_invalid_snippets_only_one_folder_json.tf"),
						ExpectError: regexp.MustCompile(`Error: snippets file should be under 'property-snippets' folder with .json extension. Invalid file: property-snippet/template_invalid_json.json`),
					},
				},
			})
		})
	})

}

func TestFormatValue(t *testing.T) {
	tests := map[string]struct {
		given     interface{}
		expected  interface{}
		withError bool
	}{
		"string": {
			given:    "test",
			expected: `"test"`,
		},
		"map": {
			given:    map[string]interface{}{"string": "value", "num": 1, "map": map[string]interface{}{"bool": true}},
			expected: `{"map":{"bool":true},"num":1,"string":"value"}`,
		},
		"number": {
			given:    1.23,
			expected: 1.23,
		},
		"boolean": {
			given:    true,
			expected: true,
		},
		"unmarshalable map": {
			given:     map[string]interface{}{"f": func() {}},
			withError: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := formatValue(test.given)
			if test.withError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

func TestGetValuesFromMap(t *testing.T) {
	variablesPath := "testData/TestDSRulesTemplate/variables"
	tests := map[string]struct {
		definitionsFile string
		valuesFile      string
		expected        map[string]interface{}
		withError       error
	}{
		"definitions and values passed, with values overwrite": {
			definitionsFile: "simple_definitions.json",
			valuesFile:      "simple_values.json",
			expected: map[string]interface{}{
				"testString": `"test 2"`,
				"testJSON":   `{"abc":"bca"}`,
				"testNumber": "null",
			},
		},
		"values not passed, take defaults": {
			definitionsFile: "simple_definitions.json",
			expected: map[string]interface{}{
				"testString": `"test"`,
				"testJSON":   `{"abc":"abc"}`,
				"testNumber": "null",
			},
		},
		"definitions file not found": {
			definitionsFile: "not_existing.json",
			withError:       ErrReadFile,
		},
		"values file not found": {
			definitionsFile: "simple_definitions.json",
			valuesFile:      "not_existing.json",
			withError:       ErrReadFile,
		},
		"invalid definitions schema": {
			definitionsFile: "invalid_definition_schema.json",
			withError:       ErrUnmarshal,
		},
		"invalid values json": {
			definitionsFile: "simple_definitions.json",
			valuesFile:      "invalid_values.json",
			withError:       ErrUnmarshal,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defPath := fmt.Sprintf("%s/%s", variablesPath, test.definitionsFile)
			var valPath string
			if test.valuesFile != "" {
				valPath = fmt.Sprintf("%s/%s", variablesPath, test.valuesFile)
			}
			res, err := getVarsFromFile(
				defPath,
				valPath)
			if test.withError != nil {
				assert.True(t, errors.Is(err, test.withError), "expected: %s; got: %s", test.expected, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

func TestConvertToTypedMap(t *testing.T) {
	tests := map[string]struct {
		givenVars []interface{}
		expected  map[string]interface{}
		withError error
	}{
		"valid map": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "testString", "type": "string", "value": "test"},
				map[string]interface{}{"name": "testNum", "type": "number", "value": "1.23"},
				map[string]interface{}{"name": "testJSONArray", "type": "jsonArray", "value": `[1, "two", false]`},
				map[string]interface{}{"name": "testJSONBlock", "type": "jsonBlock", "value": `{"abc": "cba", "number":1}`},
				map[string]interface{}{"name": "testBool", "type": "bool", "value": "true"},
			},
			expected: map[string]interface{}{
				"testString":    `"test"`,
				"testNum":       1.23,
				"testJSONArray": `[1, "two", false]`,
				"testJSONBlock": `{"abc": "cba", "number":1}`,
				"testBool":      true,
			},
		},
		"invalid values slice": {
			givenVars: []interface{}{"test"},
			withError: tools.ErrInvalidType,
		},
		"missing 'name' in map": {
			givenVars: []interface{}{
				map[string]interface{}{"type": "string", "value": "test"},
			},
			withError: tools.ErrNotFound,
		},
		"missing 'type' in map": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "testString", "value": "test"},
			},
			withError: tools.ErrNotFound,
		},
		"missing 'value' in map": {
			givenVars: []interface{}{
				map[string]interface{}{"type": "string", "name": "test"},
			},
			withError: tools.ErrNotFound,
		},
		"'name' is of invalid type": {
			givenVars: []interface{}{
				map[string]interface{}{"name": 123, "type": "string", "value": "test"},
			},
			withError: tools.ErrInvalidType,
		},
		"'type' is of invalid type": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "test", "type": 123, "value": "test"},
			},
			withError: tools.ErrInvalidType,
		},
		"'value' is of invalid type": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "test", "type": "string", "value": 123},
			},
			withError: tools.ErrInvalidType,
		},
		"jsonArray has invalid json": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "testJSON", "type": "jsonArray", "value": "abc"},
			},
			withError: ErrUnmarshal,
		},
		"jsonBlock has invalid json": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "testJSON", "type": "jsonBlock", "value": "abc"},
			},
			withError: ErrUnmarshal,
		},
		"number is invalid": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "test", "type": "number", "value": "abc"},
			},
			withError: tools.ErrInvalidType,
		},
		"bool has invalid json": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "test", "type": "bool", "value": "abc"},
			},
			withError: tools.ErrInvalidType,
		},
		"unknown type passed": {
			givenVars: []interface{}{
				map[string]interface{}{"name": "test", "type": "unknown", "value": "abc"},
			},
			withError: ErrUnknownType,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := convertToTypedMap(test.givenVars)
			if test.withError != nil {
				assert.True(t, errors.Is(err, test.withError), "want: %s; got: %s", test.withError, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

func TestConvertToTemplate(t *testing.T) {
	templates := "testData/TestDSRulesTemplate/rules/property-snippets"
	templatesOut := "testData/TestDSRulesTemplate/output"
	tests := map[string]struct {
		givenFile    string
		expectedFile string
		withError    error
	}{
		"valid conversion": {
			givenFile:    "template_in.json",
			expectedFile: "template_out.json",
		},
		"plain JSON passed": {
			givenFile:    "plain_json.json",
			expectedFile: "plain_json.json",
		},
		"invalid path": {
			givenFile: "invalid.json",
			withError: ErrReadFile,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := convertToTemplate(fmt.Sprintf("%s/%s", templates, test.givenFile))
			if test.withError != nil {
				assert.True(t, errors.Is(err, test.withError), "want: %s; got: %s", test.withError, err)
				return
			}
			require.NoError(t, err)
			expected := loadFixtureString(fmt.Sprintf("%s/%s", templatesOut, test.expectedFile))
			assert.Equal(t, expected, res)
		})
	}
}
