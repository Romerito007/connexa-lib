// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package whatsmeow

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"go.mau.fi/whatsmeow/socket"
	"go.mau.fi/whatsmeow/store"
)

var clientVersionRegex = regexp.MustCompile(`"client_revision":(\d+),`)

// GetLatestVersion returns the latest version number from web.whatsapp.com.
//
// After fetching, you can update the version to use using store.SetWAVersion, e.g.
//
//	latestVer, err := GetLatestVersion(nil)
//	if err != nil {
//		return err
//	}
//	store.SetWAVersion(*latestVer)
func GetLatestVersion(httpClient *http.Client) (*store.WAVersionContainer, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	req, err := http.NewRequest(http.MethodGet, socket.Origin, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header = getDynamicHeaders(store.DeviceProps, store.BaseClientPayload.UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response with status %d: %s", resp.StatusCode, data)
	}
	match := clientVersionRegex.FindSubmatch(data)
	if len(match) == 0 {
		return nil, fmt.Errorf("version number not found")
	}
	parsedVer, err := strconv.ParseInt(string(match[1]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version number: %w", err)
	}
	return &store.WAVersionContainer{2, 3000, uint32(parsedVer)}, nil
}
