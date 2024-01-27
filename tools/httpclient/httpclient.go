package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	SetBaseUrl(baseUrl string) HTTPClient
	SetAuthToken(token string) HTTPClient
	Bearer(token string) HTTPClient
	ToJson(any) HTTPClient
	ToBytesBuffer(*[]byte) HTTPClient
	ToPlainText(*string) HTTPClient
	Error(any) HTTPClient
	ContentType(string) HTTPClient
	Get(path string) (*http.Response, error)
	Post(path string, body any) (*http.Response, error)
	Put(path string, body any) (*http.Response, error)
	Delete(path string) (*http.Response, error)
}

type HTTPClientImpl struct {
	BaseURL   string
	AuthToken string

	HTTPClient *http.Client

	toRespBody    any
	bytesRespBody *[]byte
	toTextResp    *string
	errorResp     any
	contentType   string
}

func NewHTTPClient(baseURL, authToken string) HTTPClient {
	return &HTTPClientImpl{
		BaseURL:     baseURL,
		AuthToken:   authToken,
		HTTPClient:  &http.Client{},
		contentType: "application/json",
	}
}

func (c *HTTPClientImpl) SetBaseUrl(baseURL string) HTTPClient {
	c.BaseURL = baseURL
	return c
}

func (c *HTTPClientImpl) SetAuthToken(token string) HTTPClient {
	c.AuthToken = token
	return c
}

func (c *HTTPClientImpl) Bearer(token string) HTTPClient {
	c.AuthToken = fmt.Sprintf("Bearer %s", token)
	return c
}

func (c *HTTPClientImpl) ToJson(respBody any) HTTPClient {
	c.toRespBody = respBody
	return c
}

func (c *HTTPClientImpl) ToBytesBuffer(byteArr *[]byte) HTTPClient {
	c.bytesRespBody = byteArr
	return c
}

func (c *HTTPClientImpl) ToPlainText(textResponse *string) HTTPClient {
	c.toTextResp = textResponse
	return c
}

func (c *HTTPClientImpl) ContentType(contentType string) HTTPClient {
	c.contentType = contentType
	return c
}

func (c *HTTPClientImpl) Error(errVar any) HTTPClient {
	c.errorResp = errVar
	return c
}

func (c *HTTPClientImpl) prepareRequest(method, path string, body any) (*http.Request, error) {
	url := c.BaseURL + path

	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", c.contentType)
	req.Header.Set("Authorization", c.AuthToken)

	return req, nil
}

func (c *HTTPClientImpl) handleResponseBody(respStatusCode int, respBody []byte) error {
	if respBody == nil {
		return nil
	}

	if respStatusCode > 399 && c.errorResp != nil {
		err := json.Unmarshal(respBody, c.errorResp)
		if err != nil {
			return err
		}
	}

	if c.toRespBody != nil {
		err := json.Unmarshal(respBody, c.toRespBody)
		if err != nil {
			return err
		}
	} else if c.bytesRespBody != nil {
		c.bytesRespBody = &respBody
	} else if c.toTextResp != nil {
		strResp := string(respBody)
		*c.toTextResp = strResp
	}

	return nil
}

func (c *HTTPClientImpl) do(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer req.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}
	err = c.handleResponseBody(resp.StatusCode, respBody)

	return resp, err
}

func (c *HTTPClientImpl) Get(path string) (*http.Response, error) {
	req, err := c.prepareRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *HTTPClientImpl) Post(path string, body any) (*http.Response, error) {
	req, err := c.prepareRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *HTTPClientImpl) Put(path string, body any) (*http.Response, error) {
	req, err := c.prepareRequest(http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *HTTPClientImpl) Delete(path string) (*http.Response, error) {
	req, err := c.prepareRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}
