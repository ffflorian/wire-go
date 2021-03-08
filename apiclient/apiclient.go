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
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ffflorian/wire-go/util"
)

// APIClient is a configuration struct for the APIClient
type APIClient struct {
	AccessToken string
	Backend     string
	Cookie      *http.Cookie
	Email       string
	Password    string
	Timeout     int
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
	ClientType string `json:"clientType"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

// ClientClassification defines the type of client
type ClientClassification struct {
	Desktop   string `json:"DESKTOP"`
	LegalHold string `json:"LEGAL_HOLD"`
	Phone     string `json:"PHONE"`
	Tablet    string `json:"TABLET"`
}

// PublicClient defines a client of another user
type PublicClient struct {
	Class string `json:"class"`
	ID    string `json:"id"`
}

// ClientType defines the type of a client
type ClientType struct {
	Permanent string `json:"PERMANENT"`
	Temporary string `json:"TEMPORARY"`
}

var clientTypes = &ClientType{
	Permanent: "PERMANENT",
	Temporary: "TEMPORARY",
}

// Location defines the location of a client
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
	Time string `json:"time"`
	Type string `json:"type"`
}

// RegisteredClient defines the client of the current user
type RegisteredClient struct {
	AddedClient
	/** The cookie label */
	Cookie string `json:"cookie"`
}

type Password struct {
	Password string `json:"password"`
}

var paths = struct {
	CLIENTS string
	LOGIN   string
	USERS   string
}{
	CLIENTS: "clients",
	LOGIN:   "login",
	USERS:   "users",
}

var methods = struct {
	GET    string
	POST   string
	DELETE string
}{
	GET:    "GET",
	POST:   "POST",
	DELETE: "DELETE",
}

// New returns a new instance of the APIClient
func New(backend, email, password string, timeout int) *APIClient {
	pat := regexp.MustCompile(`https?://`)
	backendWithoutProtocol := pat.ReplaceAllString(backend, "")

	return &APIClient{
		AccessToken: "",
		Backend:     backendWithoutProtocol,
		Email:       email,
		Password:    password,
		Timeout:     timeout,
	}
}

// DeleteClient deletes a client of the current user
func (apiClient *APIClient) DeleteClient(clientID string) error {
	if apiClient.AccessToken == "" {
		return errors.New("Not logged in yet")
	}

	deleteClientData := &Password{
		Password: apiClient.Password,
	}

	urlPath := apiClient.buildURL(paths.CLIENTS, clientID)

	_, requestError := apiClient.request(methods.DELETE, urlPath, deleteClientData, true)
	if requestError != nil {
		return requestError
	}

	return nil
}

// GetClient gets a clients of the current user
func (apiClient *APIClient) GetClient(userID, clientID string) (*[]byte, error) {
	if apiClient.AccessToken == "" {
		return nil, errors.New("Not logged in yet")
	}

	urlPath := apiClient.buildURL(paths.CLIENTS, clientID)

	clients, requestError := apiClient.request(methods.GET, urlPath, nil, true)
	if requestError != nil {
		return nil, requestError
	}

	return clients, nil
}

// GetAllClients gets all clients of the current user
func (apiClient *APIClient) GetAllClients() (*[]RegisteredClient, error) {
	if apiClient.AccessToken == "" {
		return nil, errors.New("Not logged in yet")
	}

	urlPath := apiClient.buildURL(paths.CLIENTS)

	data, requestError := apiClient.request(methods.GET, urlPath, nil, true)
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

	var clientType = clientTypes.Temporary

	if permanent == true {
		clientType = clientTypes.Permanent
	}

	urlPath := apiClient.buildURL(paths.LOGIN)
	loginData := &LoginData{
		ClientType: clientType,
		Email:      apiClient.Email,
		Password:   apiClient.Password,
	}

	data, requestError := apiClient.request(methods.POST, urlPath, loginData, false)
	if requestError != nil {
		return nil, requestError
	}

	var tokenData *TokenData

	unmarshalError := json.Unmarshal(*data, &tokenData)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	apiClient.AccessToken = fmt.Sprintf("%s %s", tokenData.TokenType, tokenData.AccessToken)

	return tokenData, nil
}

func (apiClient *APIClient) buildURL(fragments ...string) string {
	path := strings.Join(fragments, "/")
	URL := &url.URL{Scheme: "https", Host: apiClient.Backend, Path: path}
	return URL.String()
}

func (apiClient *APIClient) request(method, urlPath string, payload interface{}, loginNeeded bool) (*[]byte, error) {
	timeout := time.Duration(apiClient.Timeout) * time.Millisecond

	payloadBuf := new(bytes.Buffer)

	if payload != nil {
		json.NewEncoder(payloadBuf).Encode(payload)
	}

	request, _ := http.NewRequest(method, urlPath, payloadBuf)
	request.Header.Set("Content-Type", "application/json")

	if apiClient.AccessToken != "" {
		request.Header.Set("Authorization", apiClient.AccessToken)
		fmt.Printf("Setting access token: \"%s\"\n", util.Shorten(apiClient.AccessToken, 20))
	} else if loginNeeded == true {
		return nil, errors.New("No access token saved. Not logged in?")
	}

	if apiClient.Cookie != nil {
		request.AddCookie(apiClient.Cookie)
		fmt.Printf("Setting cookie: \"%s\"\n", util.Shorten(apiClient.Cookie.String(), 20))
	} else if loginNeeded == true {
		return nil, errors.New("No zuid cookie saved. Not logged in?")
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
		if cookie.Name == "zuid" {
			fmt.Printf("Got the zuid cookie: %s\n", util.Shorten(cookie.String(), 20))
			apiClient.Cookie = cookie
			break
		}
	}

	buffer, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return nil, readError
	}

	return &buffer, nil
}
