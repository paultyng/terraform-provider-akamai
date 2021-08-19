package siteshield

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/siteshield"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiSiteShield_data_basic(t *testing.T) {
	t.Run("get SiteShield map", func(t *testing.T) {
		client := &mocksiteshield{}

		cv := siteshield.SiteShieldMapResponse{}
		expectJS := compactJSON(loadFixtureBytes("testdata/TestDSSiteShield/SiteShield.json"))
		json.Unmarshal([]byte(expectJS), &cv)

		client.On("GetSiteShieldMap",
			mock.Anything,
			siteshield.SiteShieldMapRequest{UniqueID: 1234},
		).Return(&cv, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestDSSiteShield/get_map.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.akamai_siteshield_map.test", "map_id", "1234"),
							resource.TestCheckResourceAttr("data.akamai_siteshield_map.test", "rule_name", "a;s36.akamai.net"),
							resource.TestCheckResourceAttr("data.akamai_siteshield_map.test", "acknowledged", "false"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
