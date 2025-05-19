package main

import (
	"fmt"
	"zhanghefan123/security_topology/modules/utils/remote"
)

func main() {
	client, err := remote.CreateSSHClient("zeusnet", "zeusnet123", "192.168.110.92")
	fmt.Println(err) // ssh: handshake failed: knownhosts: key is unknown
	if err == nil {
		out, _ := client.Run("ls -al")
		fmt.Println(string(out))
	}
}
