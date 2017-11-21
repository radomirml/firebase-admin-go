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

package iid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/option"

	"firebase.google.com/go/internal"

	"golang.org/x/net/context"
)

var testIIDConfig = &internal.InstanceIDConfig{
	ProjectID: "test-project",
	Opts: []option.ClientOption{
		option.WithTokenSource(&internal.MockTokenSource{AccessToken: "test-token"}),
	},
}

func TestNoProjectID(t *testing.T) {
	client, err := NewClient(context.Background(), &internal.InstanceIDConfig{})
	if client != nil || err == nil {
		t.Errorf("NewClient() = (%v, %v); want = (nil, error)", client, err)
	}
}

func TestInvalidInstanceID(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx, testIIDConfig)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteInstanceID(ctx, ""); err == nil {
		t.Errorf("DeleteInstanceID(empty) = nil; want error")
	}
}

func TestDeleteInstanceID(t *testing.T) {
	var tr *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tr = r
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer ts.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, testIIDConfig)
	if err != nil {
		t.Fatal(err)
	}
	client.endpoint = ts.URL
	if err := client.DeleteInstanceID(ctx, "test-iid"); err != nil {
		t.Errorf("DeleteInstanceID() = %v; want nil", err)
	}

	if tr == nil {
		t.Fatalf("Request = nil; want non-nil")
	}
	if tr.Method != "DELETE" {
		t.Errorf("Method = %q; want = %q", tr.Method, "DELETE")
	}
	if tr.URL.Path != "/project/test-project/instanceId/test-iid" {
		t.Errorf("Path = %q; want = %q", tr.URL.Path, "/project/test-project/instanceId/test-iid")
	}
	if h := tr.Header.Get("Authorization"); h != "Bearer test-token" {
		t.Errorf("Authorization = %q; want = %q", h, "Bearer test-token")
	}
}

func TestDeleteInstanceIDError(t *testing.T) {
	var tr *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tr = r
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer ts.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, testIIDConfig)
	if err != nil {
		t.Fatal(err)
	}
	client.endpoint = ts.URL
	if err := client.DeleteInstanceID(ctx, "test-iid"); err == nil {
		t.Errorf("DeleteInstanceID() = nil; want = error")
		return
	}

	if tr == nil {
		t.Fatalf("Request = nil; want non-nil")
	}
	if tr.Method != "DELETE" {
		t.Errorf("Method = %q; want = %q", tr.Method, "DELETE")
	}
	if tr.URL.Path != "/project/test-project/instanceId/test-iid" {
		t.Errorf("Path = %q; want = %q", tr.URL.Path, "/project/test-project/instanceId/test-iid")
	}
	if h := tr.Header.Get("Authorization"); h != "Bearer test-token" {
		t.Errorf("Authorization = %q; want = %q", h, "Bearer test-token")
	}
}

func TestDeleteInstanceIDConnectionError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do nothing
	}))
	ts.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, testIIDConfig)
	if err != nil {
		t.Fatal(err)
	}
	client.endpoint = ts.URL
	if err := client.DeleteInstanceID(ctx, "test-iid"); err == nil {
		t.Errorf("DeleteInstanceID() = nil; want = error")
		return
	}
}