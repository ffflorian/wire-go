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

func main() {
	flagContext := flags.New()

	flagContext.NewStringFlag("email", "e", "email")
	flagContext.NewStringFlag("backend", "b", "backend (default: \"staging-nginz-https.zinfra.io\"")
	flagContext.NewStringFlag("password", "p", "password")

	parseError := flagContext.Parse(os.Args...)
	if parseError != nil {
		fmt.Printf("Flag parse error: %s\n", parseError)
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

	// fmt.Println("Deleting all clients ...")
	fmt.Println("Logging in ...")

	client := apiclient.New(backend, email, password, 10000)

	DeleteAllClients(client)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// DeleteAllClients deletes all clients of the current user
func DeleteAllClients(client *apiclient.APIClient) {
	_, loginError := client.Login(false)
	if loginError != nil {
		fmt.Printf("Login error: %s\n", loginError)
		os.Exit(1)
	}

	allClients, allClientsError := client.GetAllClients()
	if allClientsError != nil {
		fmt.Printf("Error while trying to get all clients: %s\n", loginError)
		os.Exit(1)
	}

	fmt.Printf("Found %d %s.\n", len(*allClients), util.Pluralize("client", "s", len(*allClients)))

	for _, userClient := range *allClients {
		fmt.Printf("Deleting client with ID \"%s\" ...\n", userClient.ID)
		deleteError := client.DeleteClient(userClient.ID)

		if deleteError != nil {
			fmt.Printf("Error while trying to delete client with ID \"%s\": %s\n", userClient.ID, deleteError)
			os.Exit(1)
		}
	}
}
