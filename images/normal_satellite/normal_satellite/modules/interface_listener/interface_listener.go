//go:build linux

package interface_listener

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"strings"
)

// IsFrrRunning 检测 FRR 是否运行
func IsFrrRunning() bool {
	cmd := exec.Command("service", "frr", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Status of ospfd: running")
}

// AddInterfaceIntoOspf 将接口添加到 OSPF 之中
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

// RemoveInterfaceFromOSPF 将接口从 OSPF 之中移除
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

// MonitorInterface 进行接口的监听, 然后进行相应的的处理, 但是可能接口出现事件发生后, 才完成 frr 的启动
func MonitorInterface() {
	// 获取所有存在的接口
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

	fmt.Println(existingInterfaces)

	// 进行循环的监听
	for {
		select {
		// 检测更新事件
		case update := <-ch:
			// 如果是新的链路 -> 调用 ospf
			if update.Header.Type == unix.RTM_NEWLINK {
				interfaceName := update.Link.Attrs().Name
				if _, exists := existingInterfaces[interfaceName]; !exists {
					go func() {
						// 检测 frr 是否运行
						for {
							if IsFrrRunning() {
								fmt.Println(interfaceName)
								err = AddInterfaceIntoOspf(interfaceName)
								if err != nil {
									fmt.Printf("cannot add interface to ospf: %v", err)
									return
								}
								existingInterfaces[interfaceName] = struct{}{}
								break
							} else {
								continue
							}
						}
					}()
				}

				// 如果是链路被删除 -> 调用 ospf
			} else if update.Header.Type == unix.RTM_DELLINK {
				interfaceName := update.Link.Attrs().Name
				if _, exists := existingInterfaces[interfaceName]; exists {
					go func() {
						// 检测 frr 是否运行
						for {
							if IsFrrRunning() {
								fmt.Println(interfaceName)
								err = RemoveInterfaceFromOSPF(interfaceName)
								if err != nil {
									fmt.Printf("cannot remove interface from ospf: %v", err)
									return
								}
								delete(existingInterfaces, update.Link.Attrs().Name)
								break
							} else {
								continue
							}
						}
					}()
				}
			}
		}
	}
}
