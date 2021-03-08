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
	"os"

	"github.com/ffflorian/wire-go/apiclient"
	"github.com/ffflorian/wire-go/util"
	"github.com/simonleung8/flags"
)

const version = "0.0.1"

func main() {
	flagContext := flags.New()

	flagContext.NewStringFlag("backend", "b", "specify the Wire backend URL (default: \"staging-nginz-https.zinfra.io\"")
	flagContext.NewStringFlag("email", "e", "specify your Wire email address")
	flagContext.NewStringFlag("password", "p", "specify your Wire password")
	flagContext.NewBoolFlag("version", "v", "output the version number")
	flagContext.NewBoolFlag("help", "h", "display this help")

	parseError := flagContext.Parse(os.Args...)
	util.CheckError(parseError)

	if flagContext.IsSet("h") || flagContext.IsSet("help") {
		showUsage(flagContext)
		os.Exit(0)
	}

	if flagContext.IsSet("v") || flagContext.IsSet("version") {
		fmt.Println(version)
		os.Exit(0)
	}

	email := flagContext.String("e")
	backend := flagContext.String("b")
	password := flagContext.String("p")

	if email == "" {
		fmt.Println("No email set.")
		os.Exit(1)
	}

	if backend == "" {
		fmt.Println("No backend set.")
		os.Exit(1)
	}

	if password == "" {
		fmt.Println("No password set.")
		os.Exit(1)
	}

	client := apiclient.New(backend, email, password, 10000)

	var foundCommand = false

	for _, arg := range flagContext.Args() {
		if arg == "delete-all-clients" {
			deleteAllClients(client)
			foundCommand = true
			break
		}
	}

	if foundCommand == false {
		showUsage(flagContext)
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
	util.CheckError(loginError)

	allClients, allClientsError := client.GetAllClients()
	util.CheckError(allClientsError)

	fmt.Printf("Found %d %s.\n", len(*allClients), util.Pluralize("client", "s", len(*allClients)))

	for _, userClient := range *allClients {
		fmt.Printf("Deleting client with ID \"%s\" ...\n", userClient.ID)
		deleteError := client.DeleteClient(userClient.ID)

		util.CheckError(deleteError)
	}
}
