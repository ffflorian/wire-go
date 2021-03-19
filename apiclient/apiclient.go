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

	"github.com/ffflorian/go-tools/simplelogger"
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
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	UserID      string `json:"user"`
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

// SharedClient defines the base data of a client
type SharedClient struct {
	Label   string   `json:"label"`
	Lastkey PreKey   `json:"lastkey"`
	Prekeys []PreKey `json:"prekeys"`
}

// PreKey defines the PreKey needed for the Proteus protocol
type PreKey struct {
	/** The PreKey ID */
	ID int `json:"id"`
	/** The PreKey data, base64 encoded */
	Key string `json:"key"`
}

// RegisteredClient defines the client of the current user
type RegisteredClient struct {
	AddedClient
	/** The cookie label */
	Cookie string `json:"cookie"`
}

// Password defines the data to update the current user's password
type Password struct {
	Password string `json:"password"`
}

// UserAsset defines the data of a user's asset
type UserAsset struct {
	Key  string `json:"key"`
	Size string `json:"size"`
	Type string `json:"type"`
}

// SelfUpdate defines the data to update the current user
type SelfUpdate struct {
	AccentID string      `json:"accent_id"`
	Assets   []UserAsset `json:"assets"`
	Name     string      `json:"name"`
}

const (
	PathAccess  = "access"
	PathClients = "clients"
	PathLogin   = "login"
	PathLogout  = "logout"
	PathSelf    = "self"
	PathUsers   = "users"

	MethodDelete = "DELETE"
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
)

var (
	utils  = util.New("wire-go", "", "")
	logger = simplelogger.New("wire-go/apiclient", true, false)
)

// New returns a new instance of the APIClient
func New(backend, email, password string, timeout int) *APIClient {
	backendRegex := regexp.MustCompile(`https?://`)
	backendWithoutProtocol := backendRegex.ReplaceAllString(backend, "")

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
	logger.Logf("Deleting client with ID \"%s\" ...", clientID)

	if apiClient.AccessToken == "" {
		return errors.New("No access token found. Not logged in?")
	}

	deleteClientData := &Password{
		Password: apiClient.Password,
	}

	urlPath := apiClient.buildURL(PathClients, clientID)

	_, requestError := apiClient.request(MethodDelete, urlPath, deleteClientData, true)
	if requestError != nil {
		return requestError
	}

	return nil
}

// PutClient updates a client of the current user
func (apiClient *APIClient) PutClient(clientID string, updatedClient *SharedClient) error {
	logger.Logf("Updating client with ID \"%s\" ...", clientID)

	if apiClient.AccessToken == "" {
		return errors.New("No access token found. Not logged in?")
	}

	urlPath := apiClient.buildURL(PathClients, clientID)

	_, requestError := apiClient.request(MethodPut, urlPath, updatedClient, true)
	if requestError != nil {
		return requestError
	}

	return nil
}

// PutSelf updates a client of the current user
func (apiClient *APIClient) PutSelf(clientID string, updatedSelf *SelfUpdate) error {
	logger.Log("Updating user ...")

	if apiClient.AccessToken == "" {
		return errors.New("No access token found. Not logged in?")
	}

	urlPath := apiClient.buildURL(PathSelf)

	_, requestError := apiClient.request(MethodPut, urlPath, updatedSelf, true)
	if requestError != nil {
		return requestError
	}

	return nil
}

// GetClient gets a client of the current user
func (apiClient *APIClient) GetClient(clientID string) (*RegisteredClient, error) {
	logger.Logf("Getting client with ID \"%s\" ...", clientID)

	if apiClient.AccessToken == "" {
		return nil, errors.New("No access token found. Not logged in?")
	}

	urlPath := apiClient.buildURL(PathClients, clientID)

	data, requestError := apiClient.request(MethodGet, urlPath, nil, true)
	if requestError != nil {
		return nil, requestError
	}

	var client *RegisteredClient

	unmarshalError := json.Unmarshal(*data, &client)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return client, nil
}

// GetAllClients gets all clients of the current user
func (apiClient *APIClient) GetAllClients() (*[]RegisteredClient, error) {
	if apiClient.AccessToken == "" {
		return nil, errors.New("No access token found. Not logged in?")
	}

	urlPath := apiClient.buildURL(PathClients)

	data, requestError := apiClient.request(MethodGet, urlPath, nil, true)
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
	logger.Log("Logging in ...")

	var clientType = clientTypes.Temporary

	if permanent == true {
		clientType = clientTypes.Permanent
	}

	urlPath := apiClient.buildURL(PathLogin)

	loginData := &LoginData{
		ClientType: clientType,
		Email:      apiClient.Email,
		Password:   apiClient.Password,
	}

	data, requestError := apiClient.request(MethodPost, urlPath, loginData, false)
	if requestError != nil {
		return nil, requestError
	}

	var tokenData *TokenData

	unmarshalError := json.Unmarshal(*data, &tokenData)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	apiClient.AccessToken = fmt.Sprintf("%s %s", tokenData.TokenType, tokenData.AccessToken)

	logger.Logf("Got access token: \"%s\"", utils.Shorten(apiClient.AccessToken, 20))

	return tokenData, nil
}

// Logout logs the user out
func (apiClient *APIClient) Logout() error {
	logger.Log("Logging out ...")

	urlPath := apiClient.buildURL(PathAccess, PathLogout)

	_, requestError := apiClient.request(MethodPost, urlPath, nil, false)
	if requestError != nil {
		return requestError
	}

	return nil
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
		logger.Logf("Setting access token: \"%s\"", utils.Shorten(apiClient.AccessToken, 20))
	} else if loginNeeded == true {
		return nil, errors.New("No access token saved. Not logged in?")
	}

	if apiClient.Cookie != nil {
		request.AddCookie(apiClient.Cookie)
		logger.Logf("Setting cookie: \"%s\"", utils.Shorten(apiClient.Cookie.String(), 20))
	} else if loginNeeded == true {
		return nil, errors.New("No zuid cookie saved. Not logged in?")
	}

	client := &http.Client{Timeout: timeout}
	logger.Logf("Sending %s request to \"%s\" with timeout \"%s\" ...", request.Method, urlPath, timeout)

	response, requestError := client.Do(request)
	if requestError != nil {
		return nil, requestError
	}

	defer response.Body.Close()

	logger.Logf("Got response status code \"%d\"", response.StatusCode)

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("Invalid response status code")
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == "zuid" {
			logger.Logf("Got cookie: \"%s\"", utils.Shorten(cookie.String(), 20))
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
