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
	m1 := NewMockVendor(ctrl)
	m2 := NewMockVendor(ctrl)

	// when
	DeclareImplementation(m1)
	DeclareImplementation(m2)

	// then
	assert.Contains(t, vendors, m1, m2)
}

func TestCreateDataSourceWhenOneAccepts(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer func() {
		vendors = []Vendor{}
		ctrl.Finish()
	}()

	descriptor := DataSourceDescriptor{}

	m1 := NewMockVendor(ctrl)
	m1.EXPECT().Accept(gomock.Eq(&descriptor)).Return(false).Times(1)
	m1.EXPECT().CreateDataSource(gomock.Any()).Times(0)

	m2 := NewMockVendor(ctrl)
	datasource := NewMockDataSource(ctrl)
	datasource.EXPECT().Open().Times(1)
	m2.EXPECT().Accept(gomock.Eq(&descriptor)).Return(true).Times(1)
	m2.EXPECT().CreateDataSource(gomock.Eq(&descriptor)).Return(datasource, nil).Times(1)

	DeclareImplementation(m1)
	DeclareImplementation(m2)

	// when
	CreateDataSource(&DataSourceDescriptor{})
}
