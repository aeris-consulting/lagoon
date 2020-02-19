package datasource

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeclareImplementation(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer func() {
		vendors = []Vendor{}
		ctrl.Finish()
	}()
	vendor1 := NewMockVendor(ctrl)
	vendor2 := NewMockVendor(ctrl)

	// when
	DeclareImplementation(vendor1)
	DeclareImplementation(vendor2)

	// then
	assert.Contains(t, vendors, vendor1, vendor2)
}

func TestCreateDataSourceWhenOneAccepts(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer func() {
		vendors = []Vendor{}
		ctrl.Finish()
	}()

	descriptor := DataSourceDescriptor{}
	datasource := NewMockDataSource(ctrl)
	datasource.EXPECT().Open().Times(1)

	// The datasource to create is related to vendor 2 only.
	vendor1 := NewMockVendor(ctrl)
	vendor1.EXPECT().Accept(gomock.Eq(&descriptor)).Return(false).Times(1)
	vendor1.EXPECT().CreateDataSource(gomock.Any()).Times(0)

	vendor2 := NewMockVendor(ctrl)
	vendor2.EXPECT().Accept(gomock.Eq(&descriptor)).Return(true).Times(1)
	vendor2.EXPECT().CreateDataSource(gomock.Eq(&descriptor)).Return(datasource, nil).Times(1)

	DeclareImplementation(vendor1)
	DeclareImplementation(vendor2)

	// when
	result, err := CreateDataSource(&DataSourceDescriptor{})

	// then
	assert.Nil(t, err)
	assert.True(t, result == datasource)
}
