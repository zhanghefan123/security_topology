//go:build linux

package interface_listener

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
)

func AddInterfaceIntoOspf(newInterfaceName string) error {
	fmt.Printf("Adding new interface to ospf %s\n", newInterfaceName)
	networkCmd := exec.Command("vtysh",
		"-c", "configure terminal",
		"-c", fmt.Sprintf("interface %s", newInterfaceName),
		"-c", "ip ospf area 0",
		"-c", "ip ospf network point-to-point",
		"-c", "ip ospf hello-interval 5",
		"-c", "ip ospf dead-interval 20",
		"-c", "ip ospf retransmit-interval 5",
		"-c", "end")
	networkCmd.Stdout = os.Stdout
	networkCmd.Stderr = os.Stderr
	err := networkCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add new interface to ospf %s: %s", newInterfaceName, err.Error())
	} else {
		return nil
	}
}

func RemoveInterfaceFromOSPF(removedInterface string) error {
	fmt.Printf("Removing interface from ospf %s\n", removedInterface)
	networkCmd := exec.Command("vtysh",
		"-c", "configure terminal",
		"-c", fmt.Sprintf("interface %s", removedInterface),
		"-c", "no ip ospf area 0",
		"-c", "end",
	)
	networkCmd.Stdout = os.Stdout
	networkCmd.Stderr = os.Stderr
	err := networkCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to remove interface from ospf %s: %s", removedInterface, err.Error())
	} else {
		return nil
	}
}

func MonitorInterface() {
	// 所有存在的接口
	existingInterfaces := make(map[string]struct{})
	links, err := netlink.LinkList()
	if err != nil {
		fmt.Printf("cannot get link list: %v", err)
		return
	}
	for _, link := range links {
		existingInterfaces[link.Attrs().Name] = struct{}{}
	}

	ch := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	defer close(done)
	if err = netlink.LinkSubscribe(ch, done); err != nil {
		fmt.Printf("cannot subscribe to link updates: %v", err)
		return
	}

	// 进行循环的监听
	for {
		select {
		case update := <-ch:
			if update.Header.Type == unix.RTM_NEWLINK {
				interfaceName := update.Link.Attrs().Name
				if _, exists := existingInterfaces[interfaceName]; !exists {
					err = AddInterfaceIntoOspf(interfaceName)
					if err != nil {
						fmt.Printf("cannot add interface to ospf: %v", err)
						return
					}
				}
				existingInterfaces[interfaceName] = struct{}{}
			} else if update.Header.Type == unix.RTM_DELLINK {
				if _, exists := existingInterfaces[update.Link.Attrs().Name]; exists {
					err = RemoveInterfaceFromOSPF(update.Link.Attrs().Name)
					if err != nil {
						fmt.Printf("cannot remove interface from ospf: %v", err)
						return
					}
					delete(existingInterfaces, update.Link.Attrs().Name)
				}
			}
		}
	}
}
