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
	Cookies  []*http.Cookie
	Email    string
	Password string
	Timeout  int
}

// TokenData defines the data returned by the server after logging in
type TokenData struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	UserID      string `json:"user"`
	TokenType   string `json:"token_type"`
}

// LoginData defines the data sent to the server to log in
type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ClientClassification struct {
	DESKTOP    string
	LEGAL_HOLD string
	PHONE      string
	TABLET     string
}

type PublicClient struct {
	Class ClientClassification `json:"class"`
	ID    string               `json:"id"`
}

type ClientType struct {
	PERMANENT string
	TEMPORARY string
}

type Location struct {
	lat int
	lon int
}

// AddedClient defines the data returned by the server when getting a client
type AddedClient struct {
	PublicClient
	/** The IP address from which the client was registered */
	Address  string   `json:"address"`
	Label    string   `json:"label"`
	Location Location `json:"location"`
	Model    string   `json:"model"`
	/** An ISO 8601 Date string */
	Time string     `json:"time"`
	Type ClientType `json:"type"`
}

type RegisteredClient struct {
	AddedClient
	/** The cookie label */
	Cookie string `json:"cookie"`
}

type backendPaths struct {
	CLIENTS string
	LOGIN   string
	USERS   string
}

var paths = &backendPaths{
	CLIENTS: "clients",
	LOGIN:   "login",
	USERS:   "users",
}

type httpMethods struct {
	GET    string
	POST   string
	DELETE string
}

var methods = &httpMethods{
	GET:    "GET",
	POST:   "POST",
	DELETE: "DELETE",
}

// New returns a new instance of the APIClient
func New(backend string, email string, password string, timeout int) *APIClient {
	pat := regexp.MustCompile(`https?://`)
	backendWithoutProtocol := pat.ReplaceAllString(backend, "")
	var cookies []*http.Cookie

	return &APIClient{
		Backend:  backendWithoutProtocol,
		Cookies:  cookies,
		Email:    email,
		Password: password,
		Timeout:  timeout,
	}
}

// DeleteClient deletes a client of the current user
func (apiClient *APIClient) DeleteClient(clientID string) error {
	urlPath := apiClient.buildURL(paths.CLIENTS)

	_, requestError := apiClient.request(methods.DELETE, urlPath, nil)
	if requestError != nil {
		return requestError
	}

	return nil
}

// GetClient gets a clients of the current user
func (apiClient *APIClient) GetClient(userID, clientID string) (*[]byte, error) {
	urlPath := apiClient.buildURL(paths.CLIENTS, clientID)

	clients, requestError := apiClient.request(methods.GET, urlPath, nil)
	if requestError != nil {
		return nil, requestError
	}

	return clients, nil
}

// GetAllClients gets all clients of the current user
func (apiClient *APIClient) GetAllClients() (*[]RegisteredClient, error) {
	urlPath := apiClient.buildURL(paths.CLIENTS)

	data, requestError := apiClient.request(methods.GET, urlPath, nil)
	if requestError != nil {
		return nil, requestError
	}

	var clients *[]RegisteredClient

	unmarshalError := json.Unmarshal(*data, &clients)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return clients, nil
}

// Login logs the user in
func (apiClient *APIClient) Login(permanent bool) (*TokenData, error) {
	urlPath := apiClient.buildURL(paths.LOGIN)
	loginData := &LoginData{
		Email:    apiClient.Email,
		Password: apiClient.Password,
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(loginData)

	data, requestError := apiClient.request(methods.POST, urlPath, payloadBuf)
	if requestError != nil {
		return nil, requestError
	}

	fmt.Printf("Received data from server: %s\n", data)

	var tokenData *TokenData

	unmarshalError := json.Unmarshal(*data, &tokenData)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return tokenData, nil
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

	for _, cookie := range apiClient.Cookies {
		fmt.Printf("Setting a cookie named \"%s\"\n", cookie.Name)
		request.AddCookie(cookie)
	}

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

	for _, cookie := range response.Cookies() {
		fmt.Printf("Found a cookie named \"%s\"\n", cookie.Name)
	}

	apiClient.Cookies = response.Cookies()

	buffer, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return nil, readError
	}

	return &buffer, nil
}
