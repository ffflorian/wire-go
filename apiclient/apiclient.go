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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
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

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const (
	CLIENTS = "clients"
	LOGIN   = "login"
)

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

// New returns a new instance of APIClient
func New(backend string, email string, password string, timeout int) *APIClient {
	pat := regexp.MustCompile(`https?://`)
	backendWithoutProtocol := pat.ReplaceAllString(backend, "")

	return &APIClient{
		Backend:  backendWithoutProtocol,
		Email:    email,
		Password: password,
		Timeout:  timeout,
	}
}

// DeleteClient deletes a client of a user
func (apiClient *APIClient) DeleteClient(clientID string) error {
	urlPath := apiClient.buildURL(CLIENTS)

	_, requestError := apiClient.request(DELETE, urlPath, nil)
	if requestError != nil {
		return requestError
	}

	return nil
}

// GetClients gets all clients of a user
func (apiClient *APIClient) GetClients(clientID string) (*[]byte, error) {
	urlPath := apiClient.buildURL(CLIENTS, clientID)

	clients, requestError := apiClient.request(GET, urlPath, nil)
	if requestError != nil {
		return nil, requestError
	}

	return clients, nil
}

// Login logs the user in
func (apiClient *APIClient) Login(permanent bool) (*[]byte, error) {
	urlPath := apiClient.buildURL(LOGIN)
	loginData := &LoginData{
		Email:    apiClient.Email,
		Password: apiClient.Password,
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(loginData)

	data, requestError := apiClient.request(POST, urlPath, payloadBuf)
	if requestError != nil {
		return nil, requestError
	}

	fmt.Printf("Received data from server: %s\n", data)

	return data, nil
}

func (apiClient *APIClient) buildURL(fragments ...string) string {
	path := strings.Join(fragments, "/")
	URL := &url.URL{Scheme: "https", Host: apiClient.Backend, Path: path}
	return URL.String()
}

func (apiClient *APIClient) request(method string, urlPath string, payload io.Reader) (*[]byte, error) {
	timeout := time.Duration(apiClient.Timeout) * time.Millisecond
	request, _ := http.NewRequest(method, urlPath, payload)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	fmt.Printf("Sending %s request to \"%s\" with timeout \"%s\" ...\n", request.Method, urlPath, timeout)

	response, requestError := client.Do(request)
	if requestError != nil {
		return nil, requestError
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
