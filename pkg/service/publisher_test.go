package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/mock/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBackendPublisher(mockCtl *gomock.Controller, returnErr error) *interfaces.MockBackendPublisher {
	backendPublisher := interfaces.NewMockBackendPublisher(mockCtl)
	// 2秒待つので最大で2回呼ばれる想定
	// 何度かテストを回していると1回ズレる時があるので最大の設定をしている
	backendPublisher.EXPECT().FindBackendResources(gomock.Any()).MaxTimes(2).Return(returnErr)
	return backendPublisher
}

func TestRefetchResourcesWithNoError(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	backendPublisher := testBackendPublisher(mockCtl, nil)
	cfg := &config.Publisher{
		AWS: config.AWS{},
		RefetchTimer: config.RefetchTimer{
			Enabled:  true,
			Interval: 1 * time.Second,
		},
	}
	p := New(cfg, backendPublisher)
	errCh := p.RefetchResources(ctx)
	var flg = true
	for err := range errCh {
		assert.NoError(t, err)
		flg = false
		continue
	}
	require.True(t, flg)
}

func TestRefetchResourcesWithError(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	backendPublisher := testBackendPublisher(mockCtl, errors.New("Test"))
	cfg := &config.Publisher{
		AWS: config.AWS{},
		RefetchTimer: config.RefetchTimer{
			Enabled:  true,
			Interval: 1 * time.Second,
		},
	}
	p := New(cfg, backendPublisher)
	errCh := p.RefetchResources(ctx)
	var flg = true
	for err := range errCh {
		assert.Error(t, err)
		assert.Equal(t, "service.refetchResources: Test", err.Error())
		flg = false
		continue
	}
	require.False(t, flg)
}
