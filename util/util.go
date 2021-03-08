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
)

// Pluralize adds a postfix to a string
func Pluralize(text, postfix string, times int) string {
	if times == 1 {
		return text
	}
	return fmt.Sprintf("%s%s", text, postfix)
}

// Shorten returns a string with a given length and adds an ellipsis
func Shorten(text string, length int) string {
	return fmt.Sprintf("%s...", text[:length])
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
