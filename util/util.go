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

package util

import (
	"fmt"
	"os"

	"github.com/simonleung8/flags"
)

// Util is a configuration struct for the util
type Util struct {
	Description string
	FlagContext flags.FlagContext
	Name        string
	Version     string
}

// New returns a new instance of Util
func New(name string, version string, description string) *Util {
	flagContext := flags.New()

	return &Util{
		Description: description,
		FlagContext: flagContext,
		Name:        name,
		Version:     version,
	}
}

// CheckFlags checks which command line flags are set
func (util *Util) CheckFlags() {
	util.FlagContext.NewStringFlag("backend", "b", "specify the Wire backend URL (default: \"staging-nginz-https.zinfra.io\")")
	util.FlagContext.NewStringFlag("email", "e", "specify your Wire email address")
	util.FlagContext.NewStringFlag("password", "p", "specify your Wire password")
	util.FlagContext.NewBoolFlag("version", "v", "output the version number")
	util.FlagContext.NewBoolFlag("help", "h", "display this help")
	util.FlagContext.NewStringFlag("client-id", "i", "specify the client's ID (e.g. for setting its label)")
	util.FlagContext.NewStringFlag("label", "l", "specify the client's new label")

	parseError := util.FlagContext.Parse(os.Args...)
	util.CheckError(parseError, false)
}

// GetUsage returns the usage text
func (util *Util) GetUsage() string {
	return fmt.Sprintf(
		`%s

Usage:
  %s [options] [directory]

Options:
%s
Commands:
%s`,
		util.Description,
		util.Name,
		util.FlagContext.ShowUsage(2),
		util.getCommands(),
	)
}

// Pluralize adds a postfix to a string
func (util *Util) Pluralize(text, postfix string, times int) string {
	if times == 1 {
		return text
	}
	return fmt.Sprintf("%s%s", text, postfix)
}

// Shorten returns a string with a given length and adds an ellipsis
func (util *Util) Shorten(text string, length int) string {
	return fmt.Sprintf("%s...", text[:length])
}

// CheckError checks the error and if it exists, exits with exit code 1
func (util *Util) CheckError(err error, printUsage bool) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if printUsage {
			fmt.Fprintln(os.Stderr, util.GetUsage())
		}
		os.Exit(1)
	}
}

// LogAndExit logs one or more messages and exits with exit code 0
func (util *Util) LogAndExit(messages ...interface{}) {
	fmt.Println(messages...)
	os.Exit(0)
}

func (util *Util) getCommands() string {
	return `  delete-all-clients
  set-client-label
  get-all-clients`
}
