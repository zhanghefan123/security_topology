package main

import (
	"fmt"
	"os"
	"os/signal"
	"zhanghefan123/security/normal_satellite/modules/frr"
	"zhanghefan123/security/normal_satellite/modules/interface_listener"
	"zhanghefan123/security/normal_satellite/modules/srv6"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer signal.Stop(signalChan)

	// 启动流程
	// =======================================================
	PrintExitLogo()
	go interface_listener.MonitorInterface()
	srv6.Start()
	frr.Start()
	// =======================================================

	<-signalChan

	// 删除流程
	// =======================================================
	PrintRemovedLogo()
	// =======================================================
}

func PrintExitLogo() {
	fmt.Println("<------------------------------------->")
	fmt.Println("            enter ctl+c exit           ")
	fmt.Println("<------------------------------------->")
}

func PrintRemovedLogo() {
	fmt.Println("<------------------------------------->")
	fmt.Println("            satellite killed           ")
	fmt.Println("<------------------------------------->")
}
