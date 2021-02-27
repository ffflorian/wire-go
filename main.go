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

package wire

import (
	"fmt"

	"github.com/ffflorian/wire-go/apiclient"
	"github.com/simonleung8/flags"
)

func main() {
	flagContext := flags.New()

	email := flagContext.String("e")
	backend := flagContext.String("b")
	password := flagContext.String("p")

	// fmt.Println("Deleting all clients ...")
	fmt.Println("Logging in ...")

	client := apiclient.New(backend, email, password, 10000)

	_, loginError := client.Login(false)
	if loginError != nil {
		fmt.Printf("Login error: %s", loginError)
	}
}
