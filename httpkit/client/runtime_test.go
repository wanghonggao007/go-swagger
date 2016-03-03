// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-swagger/go-swagger/client"
	"github.com/go-swagger/go-swagger/httpkit"
	"github.com/go-swagger/go-swagger/strfmt"
	"github.com/stretchr/testify/assert"
)

// task This describes a task. Tasks require a content property to be set.
type task struct {

	// Completed
	Completed bool `json:"completed" xml:"completed"`

	// Content Task content can contain [GFM](https://help.github.com/articles/github-flavored-markdown/).
	Content string `json:"content" xml:"content"`

	// ID This id property is autogenerated when a task is created.
	ID int64 `json:"id" xml:"id"`
}

func TestRuntime_Concurrent(t *testing.T) {
	// test that it can make a simple request
	// and get the response for it.
	// defaults all the way down
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime)
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	hu, _ := url.Parse(server.URL)
	runtime := New(hu.Host, "/", []string{"http"})
	resCC := make(chan interface{})
	errCC := make(chan error)
	var res interface{}
	var err error

	for j := 0; j < 6; j++ {
		go func() {
			resC := make(chan interface{})
			errC := make(chan error)

			go func() {
				var resp interface{}
				var errp error
				for i := 0; i < 3; i++ {
					resp, errp = runtime.Submit(&client.Operation{
						ID:          "getTasks",
						Method:      "GET",
						PathPattern: "/",
						Params:      rwrtr,
						Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
							if response.Code() == 200 {
								var result []task
								if err := consumer.Consume(response.Body(), &result); err != nil {
									return nil, err
								}
								return result, nil
							}
							return nil, errors.New("Generic error")
						}),
					})
					<-time.After(100 * time.Millisecond)
				}
				resC <- resp
				errC <- errp
			}()
			resCC <- <-resC
			errCC <- <-errC
		}()
	}

	c := 6
	for c > 0 {
		res = <-resCC
		err = <-errCC
		c--
	}

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

func TestRuntime_Canary(t *testing.T) {
	// test that it can make a simple request
	// and get the response for it.
	// defaults all the way down
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime)
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	hu, _ := url.Parse(server.URL)
	runtime := New(hu.Host, "/", []string{"http"})
	res, err := runtime.Submit(&client.Operation{
		ID:          "getTasks",
		Method:      "GET",
		PathPattern: "/",
		Params:      rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

type tasks struct {
	Tasks []task `xml:"task"`
}

func TestRuntime_XMLCanary(t *testing.T) {
	// test that it can make a simple XML request
	// and get the response for it.
	result := tasks{
		Tasks: []task{
			{false, "task 1 content", 1},
			{false, "task 2 content", 2},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add(httpkit.HeaderContentType, httpkit.XMLMime)
		rw.WriteHeader(http.StatusOK)
		xmlgen := xml.NewEncoder(rw)
		xmlgen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	hu, _ := url.Parse(server.URL)
	runtime := New(hu.Host, "/", []string{"http"})
	res, err := runtime.Submit(&client.Operation{
		ID:          "getTasks",
		Method:      "GET",
		PathPattern: "/",
		Params:      rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result tasks
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, tasks{}, res)
		actual := res.(tasks)
		assert.EqualValues(t, result, actual)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestRuntime_CustomTransport(t *testing.T) {
	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}

	runtime := New("localhost:3245", "/", []string{"ws", "wss", "https"})
	runtime.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Scheme != "https" {
			return nil, errors.New("this was not a https request")
		}
		var resp http.Response
		resp.StatusCode = 200
		resp.Header = make(http.Header)
		resp.Header.Set("content-type", "application/json")
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		enc.Encode(result)
		resp.Body = ioutil.NopCloser(buf)
		return &resp, nil
	})

	res, err := runtime.Submit(&client.Operation{
		ID:          "getTasks",
		Method:      "GET",
		PathPattern: "/",
		Schemes:     []string{"ws", "wss", "https"},
		Params:      rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

func TestRuntime_CustomCookieJar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authenticated := false
		for _, cookie := range req.Cookies() {
			if cookie.Name == "sessionid" && cookie.Value == "abc" {
				authenticated = true
			}
		}
		if !authenticated {
			username, password, ok := req.BasicAuth()
			if ok && username == "username" && password == "password" {
				authenticated = true
				http.SetCookie(rw, &http.Cookie{Name: "sessionid", Value: "abc"})
			}
		}
		if authenticated {
			rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime)
			rw.WriteHeader(http.StatusOK)
			jsongen := json.NewEncoder(rw)
			jsongen.Encode([]task{})
		} else {
			rw.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	hu, _ := url.Parse(server.URL)
	runtime := New(hu.Host, "/", []string{"http"})
	runtime.Jar, _ = cookiejar.New(nil)

	submit := func(authInfo client.AuthInfoWriter) {
		_, err := runtime.Submit(&client.Operation{
			ID:          "getTasks",
			Method:      "GET",
			PathPattern: "/",
			Params:      rwrtr,
			AuthInfo:    authInfo,
			Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
				if response.Code() == 200 {
					return nil, nil
				}
				return nil, errors.New("Generic error")
			}),
		})

		assert.NoError(t, err)
	}

	submit(BasicAuth("username", "password"))
	submit(nil)
}

func TestRuntime_AuthCanary(t *testing.T) {
	// test that it can make a simple request
	// and get the response for it.
	// defaults all the way down
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer the-super-secret-token" {
			rw.WriteHeader(400)
			return
		}
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime)
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	hu, _ := url.Parse(server.URL)

	runtime := New(hu.Host, "/", []string{"http"})
	res, err := runtime.Submit(&client.Operation{
		ID:     "getTasks",
		Params: rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
		AuthInfo: BearerToken("the-super-secret-token"),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

func TestRuntime_PickConsumer(t *testing.T) {
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/octet-stream" {
			rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime+";charset=utf-8")
			rw.WriteHeader(400)
			return
		}
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime+";charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		req.SetBodyParam(bytes.NewBufferString("hello"))
		return nil
	})

	hu, _ := url.Parse(server.URL)
	runtime := New(hu.Host, "/", []string{"http"})
	res, err := runtime.Submit(&client.Operation{
		ID:                 "getTasks",
		Method:             "POST",
		PathPattern:        "/",
		Schemes:            []string{"http"},
		ConsumesMediaTypes: []string{"application/octet-stream"},
		Params:             rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
		AuthInfo: BearerToken("the-super-secret-token"),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

func TestRuntime_ContentTypeCanary(t *testing.T) {
	// test that it can make a simple request
	// and get the response for it.
	// defaults all the way down
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer the-super-secret-token" {
			rw.WriteHeader(400)
			return
		}
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime+";charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	hu, _ := url.Parse(server.URL)
	runtime := New(hu.Host, "/", []string{"http"})
	res, err := runtime.Submit(&client.Operation{
		ID:          "getTasks",
		Method:      "GET",
		PathPattern: "/",
		Schemes:     []string{"http"},
		Params:      rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
		AuthInfo: BearerToken("the-super-secret-token"),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

func TestRuntime_ChunkedResponse(t *testing.T) {
	// test that it can make a simple request
	// and get the response for it.
	// defaults all the way down
	result := []task{
		{false, "task 1 content", 1},
		{false, "task 2 content", 2},
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer the-super-secret-token" {
			rw.WriteHeader(400)
			return
		}
		rw.Header().Add(httpkit.HeaderTransferEncoding, "chunked")
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime+";charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		jsongen := json.NewEncoder(rw)
		jsongen.Encode(result)
	}))
	defer server.Close()

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	//specDoc, err := spec.Load("../../fixtures/codegen/todolist.simple.yml")
	hu, _ := url.Parse(server.URL)

	runtime := New(hu.Host, "/", []string{"http"})
	res, err := runtime.Submit(&client.Operation{
		ID:          "getTasks",
		Method:      "GET",
		PathPattern: "/",
		Schemes:     []string{"http"},
		Params:      rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if response.Code() == 200 {
				var result []task
				if err := consumer.Consume(response.Body(), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
			return nil, errors.New("Generic error")
		}),
		AuthInfo: BearerToken("the-super-secret-token"),
	})

	if assert.NoError(t, err) {
		assert.IsType(t, []task{}, res)
		actual := res.([]task)
		assert.EqualValues(t, result, actual)
	}
}

func TestRuntime_OverrideScheme(t *testing.T) {
	runtime := New("", "/", []string{"https"})
	sch := runtime.pickScheme([]string{"http"})
	assert.Equal(t, "https", sch)
}

func TestRuntime_PreserveTrailingSlash(t *testing.T) {
	var redirected bool

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add(httpkit.HeaderContentType, httpkit.JSONMime+";charset=utf-8")

		if req.URL.Path == "/api/tasks" {
			redirected = true
			return
		}
		if req.URL.Path == "/api/tasks/" {
			rw.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	hu, _ := url.Parse(server.URL)

	runtime := New(hu.Host, "/", []string{"http"})

	rwrtr := client.RequestWriterFunc(func(req client.Request, _ strfmt.Registry) error {
		return nil
	})

	_, err := runtime.Submit(&client.Operation{
		ID:          "getTasks",
		Method:      "GET",
		PathPattern: "/api/tasks/",
		Params:      rwrtr,
		Reader: client.ResponseReaderFunc(func(response client.Response, consumer httpkit.Consumer) (interface{}, error) {
			if redirected {
				return nil, errors.New("expected Submit to preserve trailing slashes - this caused a redirect")
			}
			if response.Code() == http.StatusOK {
				return nil, nil
			}
			return nil, errors.New("Generic error")
		}),
	})

	assert.NoError(t, err)
}
