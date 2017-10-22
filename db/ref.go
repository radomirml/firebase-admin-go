// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Ref struct {
	Key  string
	Path string

	client *Client
	segs   []string
}

func (r *Ref) Parent() *Ref {
	l := len(r.segs)
	if l > 0 {
		path := strings.Join(r.segs[:l-1], "/")
		parent, _ := r.client.NewRef(path)
		return parent
	}
	return nil
}

func (r *Ref) Child(path string) (*Ref, error) {
	if strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("child path must not start with %q", "/")
	}
	fp := fmt.Sprintf("%s/%s", r.Path, path)
	return r.client.NewRef(fp)
}

func (r *Ref) Get(v interface{}) error {
	resp, err := r.client.send(&request{Method: "GET", Path: r.Path})
	if err != nil {
		return err
	}
	return resp.CheckAndParse(http.StatusOK, v)
}

func (r *Ref) GetWithETag(v interface{}) (string, error) {
	resp, err := r.client.send(&request{
		Method: "GET",
		Path:   r.Path,
		Header: map[string]string{"X-Firebase-ETag": "true"},
	})
	if err != nil {
		return "", err
	} else if err := resp.CheckAndParse(http.StatusOK, v); err != nil {
		return "", err
	}
	return resp.Header.Get("Etag"), nil
}

func (r *Ref) Set(v interface{}) error {
	resp, err := r.client.send(&request{
		Method: "PUT",
		Path:   r.Path,
		Body:   v,
		Query:  map[string]string{"print": "silent"},
	})
	if err != nil {
		return err
	}
	return resp.CheckStatus(http.StatusNoContent)
}

func (r *Ref) SetIfUnchanged(etag string, v interface{}) (bool, error) {
	ok, _, err := r.compareAndSet(etag, v)
	return ok, err
}

func (r *Ref) Push(v interface{}) (*Ref, error) {
	resp, err := r.client.send(&request{
		Method: "POST",
		Path:   r.Path,
		Body:   v,
	})
	if err != nil {
		return nil, err
	}
	var d struct {
		Name string `json:"name"`
	}
	if err := resp.CheckAndParse(http.StatusOK, &d); err != nil {
		return nil, err
	}
	return r.Child(d.Name)
}

func (r *Ref) Update(v map[string]interface{}) error {
	if len(v) == 0 {
		return fmt.Errorf("value argument must be a non-empty map")
	}
	resp, err := r.client.send(&request{
		Method: "PATCH",
		Path:   r.Path,
		Body:   v,
		Query:  map[string]string{"print": "silent"},
	})
	if err != nil {
		return err
	}
	return resp.CheckStatus(http.StatusNoContent)
}

type UpdateFn func(interface{}) (interface{}, error)

func (r *Ref) Transaction(fn UpdateFn) error {
	var curr interface{}
	etag, err := r.GetWithETag(&curr)
	if err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		new, err := fn(curr)
		if err != nil {
			return err
		}

		ok, b, err := r.compareAndSet(etag, new)
		if err != nil {
			return err
		} else if ok {
			break
		} else if err := json.Unmarshal(b, &curr); err != nil {
			return err
		}
	}
	return nil
}

func (r *Ref) Remove() error {
	resp, err := r.client.send(&request{
		Method: "DELETE",
		Path:   r.Path,
	})
	if err != nil {
		return err
	}
	return resp.CheckStatus(http.StatusOK)
}

func (r *Ref) compareAndSet(etag string, new interface{}) (bool, []byte, error) {
	resp, err := r.client.send(&request{
		Method: "PUT",
		Path:   r.Path,
		Body:   new,
		Header: map[string]string{"If-Match": etag},
	})
	if err != nil {
		return false, nil, err
	}
	if resp.Status == http.StatusPreconditionFailed {
		return false, resp.Body, nil
	} else if err := resp.CheckStatus(http.StatusOK); err != nil {
		return false, nil, err
	}
	return true, nil, nil
}
