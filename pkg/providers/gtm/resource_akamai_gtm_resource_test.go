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

func TestResGTMResource(t *testing.T) {

	t.Run("create resource", func(t *testing.T) {
		client := &gtm.Mock{}

		getCall := client.On("GetResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(nil, &gtm.Error{
			StatusCode: http.StatusNotFound,
		}).Twice()

		resp := rsrc
		client.On("CreateResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.CreateResourceRequest"),
		).Return(&gtm.CreateResourceResponse{
			Resource: rsrcCreate.Resource,
			Status:   rsrcCreate.Status,
		}, nil).Run(func(args mock.Arguments) {
			getCall.ReturnArguments = mock.Arguments{&resp, nil}
		})

		client.On("GetResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(&resp, nil).Times(3)

		client.On("GetDomainStatus",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetDomainStatusRequest"),
		).Return(getDomainStatusResponseStatus, nil)

		client.On("UpdateResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.UpdateResourceRequest"),
		).Return(updateResourceResponseStatus, nil)

		client.On("GetResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(&rsrcUpdate, nil).Times(3)

		client.On("DeleteResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.DeleteResourceRequest"),
		).Return(deleteResourceResponseStatus, nil)

		dataSourceName := "akamai_gtm_resource.tfexample_resource_1"

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResGtmResource/create_basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "name", "tfexample_resource_1"),
							resource.TestCheckResourceAttr(dataSourceName, "aggregation_type", "latest"),
						),
					},
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResGtmResource/update_basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "name", "tfexample_resource_1"),
							resource.TestCheckResourceAttr(dataSourceName, "aggregation_type", "latest"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create resource, remove outside of terraform, expect non-empty plan", func(t *testing.T) {
		client := &gtm.Mock{}

		getCall := client.On("GetResource",
			mock.Anything,
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(nil, &gtm.Error{
			StatusCode: http.StatusNotFound,
		}).Once()

		resp := rsrc
		client.On("CreateResource",
			mock.Anything,
			mock.AnythingOfType("gtm.CreateResourceRequest"),
		).Return(&gtm.CreateResourceResponse{
			Resource: rsrcCreate.Resource,
			Status:   rsrcCreate.Status,
		}, nil).Run(func(args mock.Arguments) {
			getCall.ReturnArguments = mock.Arguments{&resp, nil}
		}).Once()

		client.On("GetResource",
			mock.Anything,
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(&resp, nil).Twice()

		// Mock that the resource was deleted outside terraform
		client.On("GetResource",
			mock.Anything,
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(nil, gtm.ErrNotFound).Once()

		// For terraform test framework, we need to mock GetResource as it would actually exist before deletion
		client.On("GetResource",
			mock.Anything,
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(&resp, nil).Once()

		client.On("DeleteResource",
			mock.Anything,
			mock.AnythingOfType("gtm.DeleteResourceRequest"),
		).Return(deleteResourceResponseStatus, nil).Once()

		dataSourceName := "akamai_gtm_resource.tfexample_resource_1"

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResGtmResource/create_basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "name", "tfexample_resource_1"),
							resource.TestCheckResourceAttr(dataSourceName, "aggregation_type", "latest"),
						),
					},
					{
						Config:             testutils.LoadFixtureString(t, "testdata/TestResGtmResource/create_basic.tf"),
						ExpectNonEmptyPlan: true,
						PlanOnly:           true,
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create resource failed", func(t *testing.T) {
		client := &gtm.Mock{}

		client.On("GetResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(nil, &gtm.Error{
			StatusCode: http.StatusNotFound,
		}).Once()

		client.On("CreateResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.CreateResourceRequest"),
		).Return(nil, &gtm.Error{
			StatusCode: http.StatusBadRequest,
		})

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResGtmResource/create_basic.tf"),
						ExpectError: regexp.MustCompile("Resource Create failed"),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create resource failed - resource already exists", func(t *testing.T) {
		client := &gtm.Mock{}

		client.On("GetResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(&rsrc, nil).Once()

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResGtmResource/create_basic.tf"),
						ExpectError: regexp.MustCompile("resource already exists error"),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("create resource denied", func(t *testing.T) {
		client := &gtm.Mock{}

		client.On("GetResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.GetResourceRequest"),
		).Return(nil, &gtm.Error{
			StatusCode: http.StatusNotFound,
		}).Once()

		dr := gtm.CreateResourceResponse{}
		dr.Resource = rsrcCreate.Resource
		dr.Status = &deniedResponseStatus
		client.On("CreateResource",
			mock.Anything, // ctx is irrelevant for this test
			mock.AnythingOfType("gtm.CreateResourceRequest"),
		).Return(&dr, nil)

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResGtmResource/create_basic.tf"),
						ExpectError: regexp.MustCompile("Request could not be completed. Invalid credentials."),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})
}

func TestGTMResourceOrder(t *testing.T) {
	tests := map[string]struct {
		client        *gtm.Mock
		pathForCreate string
		pathForUpdate string
		nonEmptyPlan  bool
		planOnly      bool
	}{
		"reordered `load_servers` - no diff": {
			client:        getGTMResourceMocks(),
			pathForCreate: "testdata/TestResGtmResource/order/create.tf",
			pathForUpdate: "testdata/TestResGtmResource/order/load_servers/reorder.tf",
			nonEmptyPlan:  false,
			planOnly:      true,
		},
		"reordered `resource_instance` - no diff": {
			client:        getGTMResourceMocks(),
			pathForCreate: "testdata/TestResGtmResource/order/create.tf",
			pathForUpdate: "testdata/TestResGtmResource/order/resource_instance/reorder.tf",
			nonEmptyPlan:  false,
			planOnly:      true,
		},
		"reordered `resource_instance` and `load_servers` - no diff": {
			client:        getGTMResourceMocks(),
			pathForCreate: "testdata/TestResGtmResource/order/create.tf",
			pathForUpdate: "testdata/TestResGtmResource/order/reorder_resource_instance_load_servers.tf",
			nonEmptyPlan:  false,
			planOnly:      true,
		},
		"change `name` attribute - diff only for `name`": {
			client:        getGTMResourceMocks(),
			pathForCreate: "testdata/TestResGtmResource/order/create.tf",
			pathForUpdate: "testdata/TestResGtmResource/order/update_name.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
		"reorder and change in `load_servers` - diff only for `load_servers`": {
			client:        getGTMResourceMocks(),
			pathForCreate: "testdata/TestResGtmResource/order/create.tf",
			pathForUpdate: "testdata/TestResGtmResource/order/load_servers/reorder_and_update.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
		"reorder resource_instance and change in `load_servers` - messy diff": {
			client:        getGTMResourceMocks(),
			pathForCreate: "testdata/TestResGtmResource/order/create.tf",
			pathForUpdate: "testdata/TestResGtmResource/order/resource_instance/reorder_and_update_load_servers.tf",
			nonEmptyPlan:  true, // change to false to see diff
			planOnly:      true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			useClient(test.client, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					IsUnitTest:               true,
					Steps: []resource.TestStep{
						{
							Config: testutils.LoadFixtureString(t, test.pathForCreate),
						},
						{
							Config:             testutils.LoadFixtureString(t, test.pathForUpdate),
							PlanOnly:           test.planOnly,
							ExpectNonEmptyPlan: test.nonEmptyPlan,
						},
					},
				})
			})
			test.client.AssertExpectations(t)
		})
	}
}

func TestResGTMResourceImport(t *testing.T) {
	tests := map[string]struct {
		domainName   string
		resourceName string
		init         func(*gtm.Mock)
		expectError  *regexp.Regexp
		stateCheck   resource.ImportStateCheckFunc
	}{
		"happy path - import": {
			domainName:   "test_domain",
			resourceName: "tfexample_resource_1",
			init: func(m *gtm.Mock) {
				// Read
				importedResource := gtm.GetResourceResponse(*getImportedResource())
				mockGetResource(m, &importedResource, nil).Times(2)
			},
			stateCheck: test.NewImportChecker().
				CheckEqual("domain", "test_domain").
				CheckEqual("name", "tfexample_resource_1").
				CheckEqual("type", "XML load object via HTTP").
				CheckEqual("host_header", "test host").
				CheckEqual("least_squares_decay", "1").
				CheckEqual("description", "test description").
				CheckEqual("leader_string", "test string").
				CheckEqual("constrained_property", "test property").
				CheckEqual("aggregation_type", "latest").
				CheckEqual("load_imbalance_percentage", "1").
				CheckEqual("upper_bound", "5").
				CheckEqual("max_u_multiplicative_increment", "10").
				CheckEqual("decay_rate", "1").
				CheckEqual("resource_instance.0.datacenter_id", "3131").
				CheckEqual("resource_instance.0.use_default_load_object", "false").
				CheckEqual("resource_instance.0.load_object", "/test1").
				CheckEqual("resource_instance.0.load_object_port", "80").
				CheckEqual("resource_instance.0.load_servers.0", "1.2.3.4").
				CheckEqual("resource_instance.0.load_servers.1", "1.2.3.5").
				CheckEqual("resource_instance.0.load_servers.2", "1.2.3.6").
				CheckEqual("wait_on_complete", "true").Build(),
		},
		"expect error - no domain name, invalid import ID": {
			domainName:   "",
			resourceName: "tfexample_resource_1",
			expectError:  regexp.MustCompile(`Error: invalid resource ID: :tfexample_resource_1`),
		},
		"expect error - no map name, invalid import ID": {
			domainName:   "test_domain",
			resourceName: "",
			expectError:  regexp.MustCompile(`Error: invalid resource ID: test_domain:`),
		},
		"expect error - read": {
			domainName:   "test_domain",
			resourceName: "tfexample_resource_1",
			init: func(m *gtm.Mock) {
				// Read - error
				mockGetResource(m, nil, fmt.Errorf("get failed")).Times(1)
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
							ImportStateId:    fmt.Sprintf("%s:%s", tc.domainName, tc.resourceName),
							ImportState:      true,
							ResourceName:     "akamai_gtm_resource.test",
							Config:           testutils.LoadFixtureString(t, "testdata/TestResGtmResource/import_basic.tf"),
							ExpectError:      tc.expectError,
						},
					},
				})
			})
			client.AssertExpectations(t)
		})
	}
}

// getGTMResourceMocks mocks creation and deletion calls for the gtm_resource
func getGTMResourceMocks() *gtm.Mock {
	client := &gtm.Mock{}

	mockGetResource := client.On("GetResource",
		mock.Anything, // ctx is irrelevant for this test
		mock.AnythingOfType("gtm.GetResourceRequest"),
	).Return(nil, &gtm.Error{
		StatusCode: http.StatusNotFound,
	})

	resp := resourceForOrderTests
	client.On("CreateResource",
		mock.Anything, // ctx is irrelevant for this test
		mock.AnythingOfType("gtm.CreateResourceRequest"),
	).Return(&gtm.CreateResourceResponse{
		Resource: rsrcCreate.Resource,
		Status:   rsrcCreate.Status,
	}, nil).Run(func(args mock.Arguments) {
		mockGetResource.ReturnArguments = mock.Arguments{&resp, nil}
	})

	client.On("DeleteResource",
		mock.Anything, // ctx is irrelevant for this test
		mock.AnythingOfType("gtm.DeleteResourceRequest"),
	).Return(deleteResourceResponseStatus, nil)

	return client
}

func mockGetResource(m *gtm.Mock, resp *gtm.GetResourceResponse, err error) *mock.Call {
	return m.On("GetResource", mock.Anything, gtm.GetResourceRequest{
		DomainName:   "test_domain",
		ResourceName: "tfexample_resource_1",
	}).Return(resp, err)
}

func getImportedResource() *gtm.Resource {
	return &gtm.Resource{
		Type:                "XML load object via HTTP",
		HostHeader:          "test host",
		LeastSquaresDecay:   1,
		Description:         "test description",
		LeaderString:        "test string",
		ConstrainedProperty: "test property",
		ResourceInstances: []gtm.ResourceInstance{
			{
				DatacenterID:         3131,
				UseDefaultLoadObject: false,
				LoadObject: gtm.LoadObject{
					LoadObject:     "/test1",
					LoadServers:    []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
					LoadObjectPort: 80,
				},
			},
		},
		AggregationType:             "latest",
		LoadImbalancePercentage:     1,
		UpperBound:                  5,
		Name:                        "tfexample_resource_1",
		MaxUMultiplicativeIncrement: 10,
		DecayRate:                   1,
	}
}

var (
	// resourceForOrderTests is a gtm.Resource structure used in testing the order of resource_instance
	resourceForOrderTests = gtm.GetResourceResponse{
		Name:            "tfexample_resource_1",
		AggregationType: "latest",
		Type:            "XML load object via HTTP",
		ResourceInstances: []gtm.ResourceInstance{
			{
				DatacenterID:         3131,
				UseDefaultLoadObject: false,
				LoadObject: gtm.LoadObject{
					LoadObject:     "/test1",
					LoadServers:    []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
					LoadObjectPort: 80,
				},
			},
			{
				DatacenterID:         3132,
				UseDefaultLoadObject: false,
				LoadObject: gtm.LoadObject{
					LoadObject:     "/test2",
					LoadServers:    []string{"1.2.3.7", "1.2.3.8", "1.2.3.9", "1.2.3.10"},
					LoadObjectPort: 80,
				},
			},
		},
	}

	rsrcCreateForOrder = gtm.CreateResourceResponse{
		Resource: &gtm.Resource{
			Name:            "tfexample_resource_1",
			AggregationType: "latest",
			Type:            "XML load object via HTTP",
			ResourceInstances: []gtm.ResourceInstance{
				{
					DatacenterID:         3131,
					UseDefaultLoadObject: false,
					LoadObject: gtm.LoadObject{
						LoadObject:     "/test1",
						LoadServers:    []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
						LoadObjectPort: 80,
					},
				},
				{
					DatacenterID:         3132,
					UseDefaultLoadObject: false,
					LoadObject: gtm.LoadObject{
						LoadObject:     "/test2",
						LoadServers:    []string{"1.2.3.7", "1.2.3.8", "1.2.3.9", "1.2.3.10"},
						LoadObjectPort: 80,
					},
				},
			},
		},
		Status: &gtm.ResponseStatus{
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
		},
	}

	rsrcCreate = gtm.CreateResourceResponse{
		Resource: &gtm.Resource{
			Name:            "tfexample_resource_1",
			AggregationType: "latest",
			Type:            "XML load object via HTTP",
			ResourceInstances: []gtm.ResourceInstance{
				{
					DatacenterID:         3131,
					UseDefaultLoadObject: false,
					LoadObject: gtm.LoadObject{
						LoadObject:     "/test1",
						LoadServers:    []string{"1.2.3.4"},
						LoadObjectPort: 80,
					},
				},
			},
		},
		Status: &gtm.ResponseStatus{
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
		},
	}

	rsrc = gtm.GetResourceResponse{
		Name:            "tfexample_resource_1",
		AggregationType: "latest",
		Type:            "XML load object via HTTP",
		ResourceInstances: []gtm.ResourceInstance{
			{
				DatacenterID:         3131,
				UseDefaultLoadObject: false,
				LoadObject: gtm.LoadObject{
					LoadObject:     "/test1",
					LoadServers:    []string{"1.2.3.4"},
					LoadObjectPort: 80,
				},
			},
		},
	}

	rsrcUpdate = gtm.GetResourceResponse{
		Name:            "tfexample_resource_1",
		AggregationType: "latest",
		Type:            "XML load object via HTTP",
		ResourceInstances: []gtm.ResourceInstance{
			{
				DatacenterID: 3132,
				LoadObject: gtm.LoadObject{
					LoadObject:     "/test2",
					LoadServers:    []string{"1.2.3.5"},
					LoadObjectPort: 80,
				},
			},
		},
	}

	updateResourceResponseStatus = &gtm.UpdateResourceResponse{
		Status: &gtm.ResponseStatus{
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
		},
	}
	deleteResourceResponseStatus = &gtm.DeleteResourceResponse{
		Status: &gtm.ResponseStatus{
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
		},
	}
)
