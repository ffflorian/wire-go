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
	"fmt"

	"github.com/ffflorian/go-tools/simplelogger"
	"github.com/ffflorian/wire-go/apiclient"
	"github.com/ffflorian/wire-go/util"
	"github.com/simonleung8/flags"
)

const (
	description = "A Wire CLI."
	name        = "wire-go"
	version     = "0.0.1"
)

var (
	logger = simplelogger.New(name, false, true)
	utils  = util.New(name, version, description)
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
		utils.LogAndExit("No email set.")
	}

	if backend == "" {
		utils.LogAndExit("No backend set.")
	}

	if password == "" {
		utils.LogAndExit("No password set.")
	}

	client := apiclient.New(backend, email, password, 10000)

	var foundCommand = false

	for _, arg := range utils.FlagContext.Args() {
		if arg == "delete-all-clients" {
			deleteAllClients(client)
			foundCommand = true
			break
		}
	}

	if foundCommand == false {
		utils.LogAndExit(utils.GetUsage())
	}
}

func showUsage(flagContext flags.FlagContext) {
	var header = `Usage: wire-cli [options] [command]

A Wire CLI.

Options:`

	fmt.Printf("%s\n%s\n", header, flagContext.ShowUsage(2))
}

func deleteAllClients(client *apiclient.APIClient) {
	fmt.Println("Logging in ...")

	_, loginError := client.Login(false)
	utils.CheckError(loginError, false)

	allClients, allClientsError := client.GetAllClients()
	utils.CheckError(allClientsError, false)

	fmt.Printf("Found %d %s.\n", len(*allClients), utils.Pluralize("client", "s", len(*allClients)))

	for _, userClient := range *allClients {
		fmt.Printf("Deleting client with ID \"%s\" ...\n", userClient.ID)
		deleteError := client.DeleteClient(userClient.ID)

		utils.CheckError(deleteError, false)
	}
}
