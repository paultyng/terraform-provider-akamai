package gtm

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/gtm"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/test"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestResGTMGeoMap(t *testing.T) {
	t.Run("create geomap", func(t *testing.T) {
		client := &gtm.Mock{}

		mockGetGeoMap(client, nil, &gtm.Error{StatusCode: http.StatusNotFound}).Once()

		mockCreateGeoMap(client, gtm.CreateGeoMapRequest{
			GeoMap:     getDefaultGeomap(),
			DomainName: testDomainName,
		}, &gtm.CreateGeoMapResponse{
			Resource: getDefaultGeomap(),
			Status:   getDefaultResponseStatus(),
		}, nil)

		resp := gtm.GetGeoMapResponse(*getDefaultGeomap())
		mockGetGeoMap(client, &resp, nil).Times(4)

		mockGetDatacenterForGeomap(client).Once()

		updateGeoMap := getDefaultUpdatedGeomap()
		mockUpdateGeoMap(client, updateGeoMap)

		mockGetDomainStatus(client, 1)

		geomapUpdate := gtm.GetGeoMapResponse(getDefaultUpdatedGeomap())
		mockGetGeoMap(client, &geomapUpdate, nil).Times(3)

		mockDeleteGeoMap(client)

		mockGetDomainStatus(client, 1)

		dataSourceName := "akamai_gtm_geomap.tfexample_geomap_1"

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/create_basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "name", "tfexample_geomap_1"),
						),
					},
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/update_basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "name", "tfexample_geomap_1"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create GEO map, remove outside of terraform, expect non-empty plan", func(t *testing.T) {
		client := &gtm.Mock{}

		mockGetGeoMap(client, nil, &gtm.Error{StatusCode: http.StatusNotFound}).Once()

		mockCreateGeoMap(client, gtm.CreateGeoMapRequest{
			GeoMap:     getDefaultGeomap(),
			DomainName: testDomainName,
		}, &gtm.CreateGeoMapResponse{
			Resource: getDefaultGeomap(),
			Status:   getDefaultResponseStatus(),
		}, nil)

		resp := gtm.GetGeoMapResponse(*getDefaultGeomap())
		mockGetGeoMap(client, &resp, nil).Twice()

		mockGetDatacenterForGeomap(client)

		// Mock that the GEOMap was deleted outside terraform
		mockGetGeoMap(client, nil, &gtm.Error{StatusCode: http.StatusNotFound}).Once()

		// For terraform test framework, we need to mock GetGEOMap as it would actually exist before deletion
		mockGetGeoMap(client, &resp, nil).Once()

		mockDeleteGeoMap(client)

		dataSourceName := "akamai_gtm_geomap.tfexample_geomap_1"

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/create_basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "name", "tfexample_geomap_1"),
						),
					},
					{
						Config:             testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/create_basic.tf"),
						ExpectNonEmptyPlan: true,
						PlanOnly:           true,
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create geomap failed", func(t *testing.T) {
		client := &gtm.Mock{}

		mockGetGeoMap(client, nil, &gtm.Error{StatusCode: http.StatusNotFound}).Once()

		mockCreateGeoMap(client, gtm.CreateGeoMapRequest{
			GeoMap:     getDefaultGeomap(),
			DomainName: testDomainName,
		}, nil, &gtm.Error{StatusCode: http.StatusBadRequest})

		mockGetDatacenterForGeomap(client).Once()

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/create_basic.tf"),
						ExpectError: regexp.MustCompile("geoMap create error"),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create geomap failed - geomap already exists", func(t *testing.T) {
		client := &gtm.Mock{}

		geomap := gtm.GetGeoMapResponse(*getDefaultGeomap())
		mockGetGeoMap(client, &geomap, nil).Once()

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/create_basic.tf"),
						ExpectError: regexp.MustCompile("geoMap already exists error"),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create geomap denied", func(t *testing.T) {
		client := &gtm.Mock{}

		mockGetGeoMap(client, nil, &gtm.Error{StatusCode: http.StatusNotFound}).Once()

		responseStatus := deniedResponseStatus
		mockCreateGeoMap(client, gtm.CreateGeoMapRequest{
			GeoMap:     getDefaultGeomap(),
			DomainName: testDomainName,
		}, &gtm.CreateGeoMapResponse{
			Resource: getDefaultGeomap(),
			Status:   &responseStatus,
		}, nil)

		mockGetDatacenterForGeomap(client).Once()

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/create_basic.tf"),
						ExpectError: regexp.MustCompile("Request could not be completed. Invalid credentials."),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})
}

func mockUpdateGeoMap(client *gtm.Mock, updateGeoMap gtm.GeoMap) *mock.Call {
	return client.On("UpdateGeoMap",
		testutils.MockContext,
		gtm.UpdateGeoMapRequest{
			GeoMap:     &updateGeoMap,
			DomainName: testDomainName,
		},
	).Return(&gtm.UpdateGeoMapResponse{
		Status: getDefaultResponseStatus(),
	}, nil).Once()
}

func TestGTMGeoMapOrder(t *testing.T) {
	tests := map[string]struct {
		pathForUpdate string
		nonEmptyPlan  bool
		planOnly      bool
	}{
		"reorder countries - no diff": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/countries/reorder.tf",
			nonEmptyPlan:  false,
			planOnly:      true,
		},
		"assignments different order - no diff": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/assignments/reorder.tf",
			nonEmptyPlan:  false,
			planOnly:      true,
		},
		"assignments and countries different order - no diff": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/reorder_assignments_and_countries.tf",
			nonEmptyPlan:  false,
			planOnly:      true,
		},
		"assignments and countries different order with updated `name` - diff only for `name`": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/update_name.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
		"assignments and countries different order with updated `domain` - diff only for `domain`": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/update_domain.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
		"assignments and countries different order with updated `wait_on_complete` - diff only for `wait_on_complete`": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/update_wait_on_complete.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
		"reordered assignments and updated countries - messy diff": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/assignments/reorder_and_update_countries.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
		"reordered assignments and updated nickname - messy diff": {
			pathForUpdate: "testdata/TestResGtmGeomap/order/assignments/reorder_and_update_nickname.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := getGeoMapOrderingTestMock()
			useClient(client, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					IsUnitTest:               true,
					Steps: []resource.TestStep{
						{
							Config: testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/order/create.tf"),
						},
						{
							Config:             testutils.LoadFixtureString(t, test.pathForUpdate),
							PlanOnly:           test.planOnly,
							ExpectNonEmptyPlan: test.nonEmptyPlan,
						},
					},
				})
			})
			client.AssertExpectations(t)
		})
	}
}

func TestResGTMGeoMapImport(t *testing.T) {
	tests := map[string]struct {
		domainName  string
		mapName     string
		init        func(*gtm.Mock)
		expectError *regexp.Regexp
		stateCheck  resource.ImportStateCheckFunc
	}{
		"happy path - import": {
			domainName: "gtm_terra_testdomain.akadns.net",
			mapName:    "tfexample_geomap_1",
			init: func(m *gtm.Mock) {
				// Read
				importedGeomap := gtm.GetGeoMapResponse(*getImportedGeoMap())
				mockGetGeoMap(m, &importedGeomap, nil).Times(2)
			},
			stateCheck: test.NewImportChecker().
				CheckEqual("domain", "gtm_terra_testdomain.akadns.net").
				CheckEqual("name", "tfexample_geomap_1").
				CheckEqual("default_datacenter.0.datacenter_id", "5400").
				CheckEqual("default_datacenter.0.nickname", "default datacenter").
				CheckEqual("assignment.0.datacenter_id", "3131").
				CheckEqual("assignment.0.nickname", "tfexample_dc_1").
				CheckEqual("assignment.0.countries.0", "GB").
				CheckEqual("wait_on_complete", "true").Build(),
		},
		"expect error - no domain name, invalid import ID": {
			domainName:  "",
			mapName:     "tfexample_geomap_1",
			expectError: regexp.MustCompile(`Error: invalid resource ID: :tfexample_geomap_1`),
		},
		"expect error - no map name, invalid import ID": {
			domainName:  "gtm_terra_testdomain.akadns.net",
			mapName:     "",
			expectError: regexp.MustCompile(`Error: invalid resource ID: gtm_terra_testdomain.akadns.net:`),
		},
		"expect error - read": {
			domainName: "gtm_terra_testdomain.akadns.net",
			mapName:    "tfexample_geomap_1",
			init: func(m *gtm.Mock) {
				// Read - error
				mockGetGeoMap(m, nil, fmt.Errorf("get failed")).Once()
			},
			expectError: regexp.MustCompile(`get failed`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &gtm.Mock{}
			if tc.init != nil {
				tc.init(client)
			}
			useClient(client, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					Steps: []resource.TestStep{
						{
							ImportStateCheck: tc.stateCheck,
							ImportStateId:    fmt.Sprintf("%s:%s", tc.domainName, tc.mapName),
							ImportState:      true,
							ResourceName:     "akamai_gtm_geomap.test",
							Config:           testutils.LoadFixtureString(t, "testdata/TestResGtmGeomap/import_basic.tf"),
							ExpectError:      tc.expectError,
						},
					},
				})
			})
			client.AssertExpectations(t)
		})
	}
}

// getGeoMapOrderingTestMock mock creation and deletion calls for gtm_geomap resource
func getGeoMapOrderingTestMock() *gtm.Mock {
	client := &gtm.Mock{}

	mockGetGeoMap(client, nil, &gtm.Error{StatusCode: http.StatusNotFound}).Once()

	mockCreateGeoMap(client, gtm.CreateGeoMapRequest{
		GeoMap:     getDiffOrderGeoMap(),
		DomainName: testDomainName,
	}, &gtm.CreateGeoMapResponse{
		Resource: getDiffOrderGeoMapForResponse(),
		Status:   getDefaultResponseStatus(),
	}, nil)

	mockGetDomainStatus(client, 1)

	resp := gtm.GetGeoMapResponse(*getDiffOrderGeoMapForResponse())
	mockGetGeoMap(client, &resp, nil).Times(4)

	mockGetDatacenterForGeomap(client).Once()

	mockGetDomainStatus(client, 1)

	mockDeleteGeoMap(client)

	return client
}

func mockGetGeoMap(client *gtm.Mock, ret *gtm.GetGeoMapResponse, err error) *mock.Call {
	return client.On("GetGeoMap",
		testutils.MockContext,
		gtm.GetGeoMapRequest{MapName: "tfexample_geomap_1", DomainName: testDomainName},
	).Return(ret, err)
}

func mockCreateGeoMap(client *gtm.Mock, request gtm.CreateGeoMapRequest, response *gtm.CreateGeoMapResponse, err error) *mock.Call {
	return client.On("CreateGeoMap", testutils.MockContext, request).Return(response, err).Once()
}

func mockGetDatacenterForGeomap(client *gtm.Mock) *mock.Call {
	return client.On("GetDatacenter",
		testutils.MockContext,
		gtm.GetDatacenterRequest{DatacenterID: 5400, DomainName: testDomainName},
	).Return(&dc, nil)
}

func mockDeleteGeoMap(client *gtm.Mock) *mock.Call {
	return client.On("DeleteGeoMap",
		testutils.MockContext,
		gtm.DeleteGeoMapRequest{MapName: "tfexample_geomap_1", DomainName: testDomainName},
	).Return(&gtm.DeleteGeoMapResponse{
		Status: getDefaultResponseStatus(),
	}, nil).Once()
}

func getDefaultDatacenter() *gtm.DatacenterBase {
	return &gtm.DatacenterBase{
		DatacenterID: 5400,
		Nickname:     "default datacenter",
	}
}

func getDefaultGeomap() *gtm.GeoMap {
	return &gtm.GeoMap{
		Name:              "tfexample_geomap_1",
		DefaultDatacenter: getDefaultDatacenter(),
		Assignments: []gtm.GeoAssignment{
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3131,
					Nickname:     "tfexample_dc_1",
				},
				Countries: []string{"GB"},
			},
		},
	}
}

func getImportedGeoMap() *gtm.GeoMap {
	return &gtm.GeoMap{
		DefaultDatacenter: &gtm.DatacenterBase{
			DatacenterID: 5400,
			Nickname:     "default datacenter",
		},
		Assignments: []gtm.GeoAssignment{
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3131,
					Nickname:     "tfexample_dc_1",
				},
				Countries: []string{"GB"},
			},
		},
		Name: "tfexample_geomap_1",
	}
}

func getDefaultUpdatedGeomap() gtm.GeoMap {
	geomap := *getDefaultGeomap()
	geomap.Assignments[0].DatacenterBase.DatacenterID = 3132
	geomap.Assignments[0].DatacenterBase.Nickname = "tfexample_dc_2"
	geomap.Assignments[0].Countries = []string{"US"}
	return geomap
}

func getDiffOrderGeoMap() *gtm.GeoMap {
	return &gtm.GeoMap{
		Name: "tfexample_geomap_1",
		DefaultDatacenter: &gtm.DatacenterBase{
			DatacenterID: 5400,
			Nickname:     "default datacenter",
		},
		Assignments: []gtm.GeoAssignment{
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3131,
					Nickname:     "tfexample_dc_1",
				},
				Countries: []string{"PL", "FR", "US", "GB"},
			},
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3132,
					Nickname:     "tfexample_dc_2",
				},
				Countries: []string{"AU", "GB"},
			},
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3133,
					Nickname:     "tfexample_dc_3",
				},
				Countries: []string{"CN", "BG", "TR", "MC", "GB"},
			},
		},
	}
}

func getDiffOrderGeoMapForResponse() *gtm.GeoMap {
	return &gtm.GeoMap{
		Name: "tfexample_geomap_1",
		DefaultDatacenter: &gtm.DatacenterBase{
			DatacenterID: 5400,
			Nickname:     "default datacenter",
		},
		Assignments: []gtm.GeoAssignment{
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3131,
					Nickname:     "tfexample_dc_1",
				},
				Countries: []string{"GB", "PL", "US", "FR"},
			},
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3132,
					Nickname:     "tfexample_dc_2",
				},
				Countries: []string{"GB", "AU"},
			},
			{
				DatacenterBase: gtm.DatacenterBase{
					DatacenterID: 3133,
					Nickname:     "tfexample_dc_3",
				},
				Countries: []string{"GB", "BG", "CN", "MC", "TR"},
			},
		},
	}
}

func getDefaultResponseStatus() *gtm.ResponseStatus {
	return &gtm.ResponseStatus{
		ChangeID: "40e36abd-bfb2-4635-9fca-62175cf17007",
		Links: []gtm.Link{
			{
				Href: "https://akab-ymtebc45gco3ypzj-apz4yxpek55y7fyv.luna.akamaiapis.net/config-gtm/v1/domains/gtmdomtest.akadns.net/status/current",
				Rel:  "self",
			},
		},
		Message:               "Current configuration has been propagated to all GTM nameservers",
		PassingValidation:     true,
		PropagationStatus:     "COMPLETE",
		PropagationStatusDate: "2019-04-25T14:54:00.000+00:00",
	}
}
