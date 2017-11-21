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

// Package iid contains functions for deleting instance IDs from Firebase projects.
package iid

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/transport"

	"firebase.google.com/go/internal"

	"golang.org/x/net/context"
)

const iidEndpoint = "https://console.firebase.google.com/v1"

// Client is the interface for the Firebase instance ID service.
type Client struct {
	endpoint string
	client   *internal.HTTPClient
	project  string
}

// NewClient creates a new instance of the Firebase instance ID Client.
//
// This function can only be invoked from within the SDK. Client applications should access the
// the instance ID service through firebase.App.
func NewClient(ctx context.Context, c *internal.InstanceIDConfig) (*Client, error) {
	if c.ProjectID == "" {
		return nil, errors.New("project id is required to access instance id client")
	}

	hc, _, err := transport.NewHTTPClient(ctx, c.Opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		endpoint: iidEndpoint,
		client:   &internal.HTTPClient{Client: hc},
		project:  c.ProjectID,
	}, nil
}

// DeleteInstanceID deletes an instance ID from Firebase.
//
// This can be used to delete an instance ID and associated user data from a Firebase project,
// pursuant to the General Data protection Regulation (GDPR).
func (c *Client) DeleteInstanceID(ctx context.Context, iid string) error {
	if iid == "" {
		return errors.New("instance id must not be empty")
	}

	url := fmt.Sprintf("%s/project/%s/instanceId/%s", c.endpoint, c.project, iid)
	resp, err := c.client.Do(ctx, &internal.Request{Method: "DELETE", URL: url})
	if err != nil {
		return err
	}
	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return err
	}
	return nil
}