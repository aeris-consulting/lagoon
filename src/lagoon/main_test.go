package main

import (
	"errors"
	"github.com/golang/mock/gomock"
	"io/ioutil"
	"lagoon/api"
	"lagoon/datasource"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirectionAtRoot(t *testing.T) {
	router := setupRouter()
	recorder := httptest.NewRecorder()
	defer recorder.Flush()

	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 301, recorder.Code)
	assert.Equal(t, "/lagoon/ui", recorder.Header().Get("Location"))
}

func TestRedirectionAtContextRoot(t *testing.T) {
	router := setupRouter()
	recorder := httptest.NewRecorder()
	defer recorder.Flush()

	req, _ := http.NewRequest("GET", contextPath, nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 301, recorder.Code)
	assert.Equal(t, "/lagoon/", recorder.Header().Get("Location"))
}

func TestRedirectionAtContextRootWithSlash(t *testing.T) {
	router := setupRouter()
	recorder := httptest.NewRecorder()
	defer recorder.Flush()

	req, _ := http.NewRequest("GET", contextPath+"/", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 301, recorder.Code)
	assert.Equal(t, "/lagoon/ui", recorder.Header().Get("Location"))
}

func TestCreateNewDataSourceListAndDelete(t *testing.T) {
	// given
	router := setupRouter()
	recorder := httptest.NewRecorder()
	ctrl := gomock.NewController(t)
	defer func() {
		ctrl.Finish()
		defer recorder.Flush()
		datasource.ClearVendors()
		api.ClearDatasources()
	}()

	// This datasource is created and deleted.
	ds1 := datasource.NewMockDataSource(ctrl)
	ds1.EXPECT().Open().Return(nil).Times(1)
	ds1.EXPECT().Close().Times(1)

	// This datasource is only created.
	ds2 := datasource.NewMockDataSource(ctrl)
	ds2.EXPECT().Open().Return(nil).Times(1)

	vendor := datasource.NewMockVendor(ctrl)
	datasource.DeclareImplementation(vendor)
	vendor.EXPECT().Accept(gomock.Any()).Return(true).Times(2)
	vendor.EXPECT().CreateDataSource(gomock.Any()).Return(ds1, nil).Times(1)
	vendor.EXPECT().CreateDataSource(gomock.Any()).Return(ds2, nil).Times(1)

	// when
	req, _ := http.NewRequest("POST", contextPath+"/datasource", strings.NewReader("{\"id\":\"my-datasource\",\"vendor\":\"mock\",\"name\":\"test-mock\",\"bootstrap\":\"any:path\"}"))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, 200, recorder.Code)
	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"dataSourceId\":\"my-datasource\"}", string(body))

	// when
	req, _ = http.NewRequest("POST", contextPath+"/datasource", strings.NewReader("{\"id\":\"my-datasource-2\",\"vendor\":\"mock\",\"name\":\"test-mock-2\",\"bootstrap\":\"any:path\"}"))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)
	// Empties the body reader.
	ioutil.ReadAll(recorder.Body)

	req, _ = http.NewRequest("GET", contextPath+"/datasource", nil)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, 200, recorder.Code)
	body, _ = ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"datasources\":[{\"id\":\"my-datasource\",\"vendor\":\"mock\",\"name\":\"test-mock\",\"description\":\"\",\"readonly\":false},{\"id\":\"my-datasource-2\",\"vendor\":\"mock\",\"name\":\"test-mock-2\",\"description\":\"\",\"readonly\":false}]}", string(body))

	// when
	req, _ = http.NewRequest("DELETE", contextPath+"/datasource/my-datasource", nil)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, 200, recorder.Code)
	body, _ = ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"message\":\"Data source was closed and removed\"}", string(body))

	// when
	req, _ = http.NewRequest("GET", contextPath+"/datasource", nil)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, 200, recorder.Code)
	body, _ = ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"datasources\":[{\"id\":\"my-datasource-2\",\"vendor\":\"mock\",\"name\":\"test-mock-2\",\"description\":\"\",\"readonly\":false}]}", string(body))
}

func TestCreateNewDataSourceWithInvalidVendor(t *testing.T) {
	// given
	router := setupRouter()
	recorder := httptest.NewRecorder()
	ctrl := gomock.NewController(t)
	defer func() {
		ctrl.Finish()
		defer recorder.Flush()
		datasource.ClearVendors()
		api.ClearDatasources()
	}()

	req, _ := http.NewRequest("POST", contextPath+"/datasource", strings.NewReader("{\"id\":\"local\",\"vendor\":\"mock\",\"name\":\"test-mock\",\"bootstrap\":\"any:path\"}"))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 400, recorder.Code)
}

func TestCreateNewDataSourceWithConnectionError(t *testing.T) {
	// given
	router := setupRouter()
	recorder := httptest.NewRecorder()
	ctrl := gomock.NewController(t)
	defer func() {
		ctrl.Finish()
		defer recorder.Flush()
		datasource.ClearVendors()
		api.ClearDatasources()
	}()

	ds := datasource.NewMockDataSource(ctrl)
	vendor := datasource.NewMockVendor(ctrl)
	datasource.DeclareImplementation(vendor)
	vendor.EXPECT().Accept(gomock.Any()).Return(true).Times(1)
	vendor.EXPECT().CreateDataSource(gomock.Any()).Return(ds, nil).Times(1)
	ds.EXPECT().Open().Return(errors.New("An error")).Times(1)

	req, _ := http.NewRequest("POST", contextPath+"/datasource", strings.NewReader("{\"id\":\"local\",\"vendor\":\"mock\",\"name\":\"test-mock\",\"bootstrap\":\"any:path\"}"))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 500, recorder.Code)
}

func TestPatchDataSource(t *testing.T) {
	// given
	router := setupRouter()
	recorder := httptest.NewRecorder()
	ctrl := gomock.NewController(t)
	defer func() {
		ctrl.Finish()
		defer recorder.Flush()
		datasource.ClearVendors()
		api.ClearDatasources()
	}()

	ds := datasource.NewMockDataSource(ctrl)
	vendor := datasource.NewMockVendor(ctrl)
	datasource.DeclareImplementation(vendor)
	vendor.EXPECT().Accept(gomock.Any()).Return(true).Times(2)
	vendor.EXPECT().CreateDataSource(gomock.Any()).Return(ds, nil).Times(2)
	ds.EXPECT().Open().Return(nil).Times(2)
	ds.EXPECT().Close().Times(1)

	req, _ := http.NewRequest("POST", contextPath+"/datasource", strings.NewReader("{\"id\":\"my-datasource\",\"vendor\":\"mock\",\"name\":\"test-mock\",\"bootstrap\":\"any:path\"}"))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	router.ServeHTTP(recorder, req)
	ioutil.ReadAll(recorder.Body)
	req, _ = http.NewRequest("GET", contextPath+"/datasource", nil)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"datasources\":[{\"id\":\"my-datasource\",\"vendor\":\"mock\",\"name\":\"test-mock\",\"description\":\"\",\"readonly\":false}]}", string(body))

	// when
	req, _ = http.NewRequest("PATCH", contextPath+"/datasource", strings.NewReader("{\"id\":\"my-datasource\",\"vendor\":\"mock\",\"name\":\"test-mock-2\",\"bootstrap\":\"any:path\",\"readonly\":true}"))
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, 200, recorder.Code)
	body, _ = ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "", string(body))

	// when
	req, _ = http.NewRequest("GET", contextPath+"/datasource", nil)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorder.Flush()
	router.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, 200, recorder.Code)
	body, _ = ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "{\"datasources\":[{\"id\":\"my-datasource\",\"vendor\":\"mock\",\"name\":\"test-mock-2\",\"description\":\"\",\"readonly\":true}]}", string(body))
}
