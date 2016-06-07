package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/howeyc/gopass"
	"github.com/joushou/qp"
	"github.com/joushou/qptools/fileserver"
)

func main() {
	jiraClient, err := jira.NewClient(nil, os.Args[1])
	if err != nil {
		fmt.Printf("Could not connect to JIRA: %v\n", err)
		return
	}

	var user, password string
	fmt.Printf("Username: ")
	_, err = fmt.Scanln(&user)
	if err == nil {

		fmt.Printf("Password: ")
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			fmt.Printf("Could not read password: %v", err)
			return
		}
		password = string(pass)

		auth := func() {
			res, err := jiraClient.Authentication.AcquireSessionCookie(user, password)
			if err != nil || res == false {
				fmt.Printf("Could not authenticate to JIRA: %v\n", err)
				return
			}
		}
		auth()

		go func() {
			t := time.NewTicker(5 * time.Minute)
			for range t.C {
				auth()
			}
		}()
	} else {
		fmt.Printf("Continuing without authentication.\n")
	}

	root, err := NewJiraDir("", 0555|qp.DMDIR, "jira", "jira", jiraClient, &JiraView{})
	if err != nil {
		fmt.Printf("Could not create JIRA view")
		return
	}

	l, err := net.Listen("tcp", ":30000")
	if err != nil {
		fmt.Printf("Could not listen: %v\n", err)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v\n", err)
			return
		}

		f := fileserver.New(conn, root, nil)
		f.Verbosity = fileserver.Quiet
		go f.Serve()
	}

}
