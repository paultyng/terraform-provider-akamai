package siteshield

import (
	"context"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/siteshield"
	"github.com/stretchr/testify/mock"
)

type mocksiteshield struct {
	mock.Mock
}

func (p *mocksiteshield) GetSiteShieldMaps(ctx context.Context) (*siteshield.GetSiteShieldMapsResponse, error) {
	args := p.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*siteshield.GetSiteShieldMapsResponse), args.Error(1)
}

func (p *mocksiteshield) GetSiteShieldMap(ctx context.Context, params siteshield.SiteShieldMapRequest) (*siteshield.SiteShieldMapResponse, error) {
	args := p.Called(ctx, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*siteshield.SiteShieldMapResponse), args.Error(1)
}

func (p *mocksiteshield) AckSiteShieldMap(ctx context.Context, params siteshield.SiteShieldMapRequest) (*siteshield.SiteShieldMapResponse, error) {
	args := p.Called(ctx, params)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*siteshield.SiteShieldMapResponse), args.Error(1)
}
