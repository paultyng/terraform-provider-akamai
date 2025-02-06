package clientlists

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/clientlists"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientList_data_single_list(t *testing.T) {
	dataSourceName := "data.akamai_clientlist_list.list"
	tests := map[string]struct {
		params       clientlists.GetClientListRequest
		config       string
		responseBody []byte
	}{
		"List": {
			params: clientlists.GetClientListRequest{
				ListID:       "123_TEST",
				IncludeItems: true,
			},
			config:       loadFixtureString("testData/TestDSClientList/get_list.tf"),
			responseBody: loadFixtureBytes("testData/TestDSClientList/ClientList.json"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &clientlists.Mock{}
			clientListResponse := clientlists.GetClientListResponse{}
			err := json.Unmarshal(test.responseBody, &clientListResponse)
			require.NoError(t, err)

			r, err := json.Marshal(clientListResponse.Items)
			require.NoError(t, err)

			buf := &bytes.Buffer{}
			err = json.Indent(buf, r, "", "  ")
			require.NoError(t, err)

			clientListResponseJSONString := buf.String()

			client.On("GetClientList",
				mock.Anything,
				test.params,
			).Return(&clientListResponse, nil)

			useClient(client, func() {
				checks := []resource.TestCheckFunc{
					resource.TestCheckResourceAttr(dataSourceName, "json", clientListResponseJSONString),
					resource.TestCheckResourceAttr(dataSourceName, "list_id", test.params.ListID),
					resource.TestCheckResourceAttr(dataSourceName, "items.#", strconv.Itoa(len(clientListResponse.Items))),
				}

				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					Steps: []resource.TestStep{
						{
							Config: test.config,
							Check:  resource.ComposeAggregateTestCheckFunc(checks...),
						},
					},
				})
			})

			client.AssertExpectations(t)
		})
	}
}
