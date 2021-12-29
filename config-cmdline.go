package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/aykevl/pwhash"
	"golang.org/x/term"
)

func (c *ConfigurationService) processCommandLine() bool {
	newAdmin := flag.String("addadmin", "", "add a new admin/password interactively to the admins list")
	newPassword := flag.String("passwd", "", "updates the password for an existing admin interactively")
	delAdmin := flag.String("rmadmin", "", "remove an existing user from the admin list")

	flag.Parse()

	switch true {
	case newAdmin != nil && *newAdmin != "":
		if _, ok := c.Values.HTTPD.Admins[*newAdmin]; ok {
			fmt.Printf("admin %s already exists!\n", *newAdmin)
			os.Exit(1)
		}
		c.Values.HTTPD.Admins[*newAdmin] = ""
		newPassword = newAdmin
		fallthrough

	case newPassword != nil && *newPassword != "":
		if _, ok := c.Values.HTTPD.Admins[*newPassword]; !ok {
			fmt.Printf("admin %s doesn't exist!\n", *newPassword)
			os.Exit(1)
		}
		c.Values.HTTPD.Admins[*newPassword] = pwhash.Hash(string(c.getPass()))
		err := c.Write()
		if err != nil {
			fmt.Printf("error updating config file [%s]", err)
			os.Exit(1)
		}
		fmt.Printf("updated configuration")
		return true

	case delAdmin != nil && *delAdmin != "":
		if _, ok := c.Values.HTTPD.Admins[*delAdmin]; !ok {
			fmt.Printf("admin %s doesn't exist!\n", *delAdmin)
			os.Exit(1)
		}
		delete(c.Values.HTTPD.Admins, *delAdmin)
		err := c.Write()
		if err != nil {
			fmt.Printf("error updating config file [%s]", err)
			os.Exit(1)
		}
		fmt.Printf("updated configuration")
		return true
	}

	return false
}

func (c *ConfigurationService) getPass() (pword []byte) {
	var err error
	fmt.Print("Password: ")
	pword, err = term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("error reading password string: [%s]\n", err)
		os.Exit(1)
	}
	fmt.Println()
	return
}
