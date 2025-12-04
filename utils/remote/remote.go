package remote

import (
	"github.com/melbahja/goph"
)

func CreateSSHClient(username, password, hostname string) (*goph.Client, error) {
	auth := goph.Password(password)
	client, err := goph.New(username, hostname, auth) // open C:\Users\zhf/.ssh/known_hosts: The system cannot find the file specified.
	if err != nil {
		return nil, err
	}
	return client, nil
}
