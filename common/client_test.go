package common

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/general-backend/goctx"
	gock "gopkg.in/h2non/gock.v1"
)

func TestHTTPDo(t *testing.T) {
	client := GetHTTPClient()
	defer gock.Off() // Flush pending mocks after test execution
	gock.InterceptClient(client)
	defer gock.RestoreClient(client)
	apDomain := "http://test.com"
	path := "/test"
	gock.New(apDomain).
		Get(path).
		Reply(200).
		JSON(map[string]string{
			"id": "123",
		})

	ctx := goctx.Background()
	req, err := http.NewRequest(http.MethodGet, "http://test.com/test", nil)
	assert.NoError(t, err)
	resp, err := HTTPDo(ctx, req)
	assert.NoError(t, err)
	// assert.NotEmpty(t, req.Header.Get("Uber-Trace-Id"))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHTTPPostForm(t *testing.T) {
	client := GetHTTPClient()
	defer gock.Off() // Flush pending mocks after test execution
	gock.InterceptClient(client)
	defer gock.RestoreClient(client)
	apDomain := "http://test.com"
	path := "/test"
	gock.New(apDomain).
		Post(path).
		Reply(200).
		JSON(map[string]string{
			"id": "123",
		})

	ctx := goctx.Background()
	resp, err := HTTPPostForm(ctx, "http://test.com/test", nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetHTTPClient(t *testing.T) {
	client := GetHTTPClient()
	assert.Equal(t, httpClient, client)
}
