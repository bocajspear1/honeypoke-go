// +build windows

package permissions

import "log"

func DropPermissions(newUser string, newGroup string) {
	log.Println("Windows has no privileges to drop...")
}
