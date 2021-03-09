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

package main

import (
	"github.com/ffflorian/go-tools/simplelogger"
	"github.com/ffflorian/wire-go/apiclient"
	"github.com/ffflorian/wire-go/util"
)

const (
	description = "A Wire CLI."
	name        = "wire-go"
	version     = "0.0.1"
)

var (
	utils  = util.New(name, version, description)
	logger = simplelogger.New(name, true, false)
)

func main() {
	utils.CheckFlags()

	email := utils.FlagContext.String("e")
	backend := utils.FlagContext.String("b")
	password := utils.FlagContext.String("p")

	if utils.FlagContext.IsSet("h") {
		utils.LogAndExit(utils.GetUsage())
	}

	if utils.FlagContext.IsSet("v") || utils.FlagContext.IsSet("version") {
		utils.LogAndExit(version)
	}

	if email == "" {
		utils.LogAndExit("Error: No email set.")
	}

	if backend == "" {
		backend = "staging-nginz-https.zinfra.io"
	}

	if password == "" {
		utils.LogAndExit("Error: No password set.")
	}

	apiClient := apiclient.New(backend, email, password, 10000)

	var foundCommand = false

	for _, arg := range utils.FlagContext.Args() {
		if arg == "delete-all-clients" {
			deleteAllClients(apiClient)
			foundCommand = true
			break
		} else if arg == "set-client-label" {
			setClientLabel(apiClient)
			foundCommand = true
			break
		} else if arg == "get-all-clients" {
			getAllClients(apiClient)
			foundCommand = true
			break
		}
	}

	if foundCommand == false {
		utils.LogAndExit(utils.GetUsage())
	}
}

func deleteAllClients(apiClient *apiclient.APIClient) {
	_, loginError := apiClient.Login(false)
	utils.CheckError(loginError, false)

	allClients, allClientsError := apiClient.GetAllClients()
	utils.CheckError(allClientsError, false)

	logger.Logf("Found %d %s.", len(*allClients), utils.Pluralize("client", "s", len(*allClients)))

	for _, userClient := range *allClients {
		deleteError := apiClient.DeleteClient(userClient.ID)

		utils.CheckError(deleteError, false)
	}

	logoutError := apiClient.Logout()
	utils.CheckError(logoutError, false)
}

func setClientLabel(apiClient *apiclient.APIClient) {
	clientID := utils.FlagContext.String("i")

	if clientID == "" {
		utils.LogAndExit("Error: No client ID set.")
	}

	label := utils.FlagContext.String("l")

	if label == "" {
		utils.LogAndExit("Error: No client label set.")
	}

	_, loginError := apiClient.Login(false)
	utils.CheckError(loginError, false)

	var updatedClient = &apiclient.SharedClient{
		Label: label,
	}

	apiClient.PutClient(clientID, updatedClient)

	logoutError := apiClient.Logout()
	utils.CheckError(logoutError, false)
}

func getAllClients(apiClient *apiclient.APIClient) {
	_, loginError := apiClient.Login(false)
	utils.CheckError(loginError, false)

	allClients, allClientsError := apiClient.GetAllClients()
	utils.CheckError(allClientsError, false)

	logger.Logf("Found %d %s.", len(*allClients), utils.Pluralize("client", "s", len(*allClients)))

	for _, userClient := range *allClients {
		logger.Log(userClient)
	}

	logoutError := apiClient.Logout()
	utils.CheckError(logoutError, false)
}
