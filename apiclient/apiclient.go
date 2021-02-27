/*
Copyright © 2021 Florian Imdahl <git@ffflorian.de>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package apiclient

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIClient is a configuration struct for the APIClient
type APIClient struct {
	Backend  string
	Email    string
	Password string
	Timeout  int
}

const (
	CLIENTS = "clients"
	LOGIN   = "clients"
)

// New returns a new instance of APIClient
func New(backend string, email string, password string, timeout int) *APIClient {
	return &APIClient{
		Backend:  backend,
		Email:    email,
		Password: password,
		Timeout:  timeout,
	}
}

// DeleteClient deletes a client of a user
func (apiClient *APIClient) DeleteClient(clientID string) error {
	urlPath := apiClient.buildURL(CLIENTS)
	request := http.Request{Method: "DELETE", URL: urlPath}

	_, requestError := apiClient.request(&request)
	if requestError != nil {
		return requestError
	}

	return nil
}

// GetClients gets all clients of a user
func (apiClient *APIClient) GetClients(clientID string) (*[]byte, error) {
	urlPath := apiClient.buildURL(CLIENTS, clientID)
	request := http.Request{Method: "GET", URL: urlPath}

	clients, requestError := apiClient.request(&request)
	if requestError != nil {
		return nil, requestError
	}

	return clients, nil
}

// Login logs the user in
func (apiClient *APIClient) Login(permanent bool) (*[]byte, error) {
	urlPath := apiClient.buildURL(LOGIN)
	request := http.Request{Method: "POST", URL: urlPath}

	data, requestError := apiClient.request(&request)
	if requestError != nil {
		return nil, requestError
	}

	fmt.Printf("Received data from server: %s\n", data)

	return data, nil
}

func (apiClient *APIClient) buildURL(fragments ...string) *url.URL {
	path := strings.Join(fragments, "/")
	return &url.URL{Scheme: "https", Host: apiClient.Backend, Path: path}
}

func (apiClient *APIClient) request(config *http.Request) (*[]byte, error) {
	timeout := time.Duration(apiClient.Timeout) * time.Millisecond
	httpClient := &http.Client{Timeout: timeout}

	fmt.Printf("Sending GET request to \"%s\" with timeout \"%s\" ...\n", config.URL.String(), timeout)

	response, responseError := httpClient.Do(config)
	if responseError != nil {
		return nil, responseError
	}

	defer response.Body.Close()

	fmt.Printf("Got response status code \"%d\"\n", response.StatusCode)

	if response.StatusCode != 200 {
		return nil, errors.New("Invalid response status code")
	}

	buffer, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return nil, readError
	}

	return &buffer, nil
}
