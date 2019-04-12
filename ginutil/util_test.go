package ginutil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/general-backend/goctx"
)

func TestParseJSON(t *testing.T) {
	body := `{"totalSize":1,"done":true,"records":[{"attributes":{"type":"Contact","url":"/services/data/v42.0/sobjects/Contact/0037F00000bbFdxQAE"},"FirstName":"test","LastName":"456","Id":"0037F00000bbFdxQAE","Email":"test@m800.com","Salutation":null,"MailingAddress":null,"Title":null,"Phone":null,"MobilePhone":null}]}`
	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
	}
	type Contact struct {
		ID string `json:"id,omitempty"`
		// required fields
		LastName   string `json:"lastName,omitempty"`
		FirstName  string `json:"firstName,omitempty"`
		Email      string `json:"email,omitempty"`
		Salutation string `json:"salutation,omitempty"`
	}
	type salesforceQueryContactResp struct {
		TotalSize int       `json:"totalSize,omitempty"`
		Done      bool      `json:"done,omitempty"`
		Records   []Contact `json:"records,omitempty"`
	}
	datas := salesforceQueryContactResp{}
	err := ParseJSON(goctx.Background(), resp.Body, &datas)
	assert.NoError(t, err)
	assert.Equal(t, 1, datas.TotalSize)
	assert.True(t, datas.Done)
	assert.Len(t, datas.Records, 1)
	assert.Equal(t, "test@m800.com", datas.Records[0].Email)
}

func TestGetStringFromIO(t *testing.T) {
	body := `{"totalSize":1,"done":true,"records":[{"attributes":{"type":"Contact","url":"/services/data/v42.0/sobjects/Contact/0037F00000bbFdxQAE"},"FirstName":"test","LastName":"456","Id":"0037F00000bbFdxQAE","Email":"test@m800.com","Salutation":null,"MailingAddress":null,"Title":null,"Phone":null,"MobilePhone":null}]}`
	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
	}

	str := GetStringFromIO(resp.Body)
	assert.True(t, len(str) > 0)
	assert.Equal(t, `{"tot`, str[:5])
}
