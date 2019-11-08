// +build !windows

package permissions

import (
	"log"
	"os/user"
	"strconv"
	"syscall"
)

import (
	//#include <unistd.h>
	//#include <errno.h>
	"C"
)

// DropPermissions drops permissions if started as root (which is normal)
func DropPermissions(newUser string, newGroup string) {
	//Got from https://stackoverflow.com/questions/41248866/golang-dropping-privileges-v1-7
	if syscall.Getuid() == 0 {

		newUserData, err := user.Lookup(newUser)
		if err != nil {
			log.Fatalln("Could not get user:", err)
			return
		}
		newGroupData, err := user.LookupGroup(newGroup)
		if err != nil {
			log.Fatalln("Could not get group: ", err)
			return
		}
		// TODO: Write error handling for int from string parsing
		uid, _ := strconv.ParseInt(newUserData.Uid, 10, 32)
		gid, _ := strconv.ParseInt(newGroupData.Gid, 10, 32)

		cerr, errno := C.setgid(C.__gid_t(gid))
		if cerr != 0 {
			log.Fatalln("Unable to set GID due to error:", errno)
		}
		cerr, errno = C.setuid(C.__uid_t(uid))
		if cerr != 0 {
			log.Fatalln("Unable to set UID due to error:", errno)
		}

	}
}
