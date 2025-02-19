package frr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// CopyFrrConfigurationFile 进行 FRR 配置文件的拷贝
func CopyFrrConfigurationFile() {
	containerName := os.Getenv("CONTAINER_NAME")

	sourceFilePath := fmt.Sprintf("/configuration/%s/route/frr.conf", containerName)

	destFilePath := "/etc/frr/frr.conf"

	cmd := exec.Command("cp", sourceFilePath, destFilePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

// StartFrr 执行启动 frr 的命令
func StartFrr() {
	cmd := exec.Command("service", "frr", "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}

// Start 拷贝 frr + 启动 frr
func Start() {
	enableFrr := os.Getenv("ENABLE_FRR")
	if enableFrr == "true" {
		fmt.Println("start frr")
		CopyFrrConfigurationFile()
		StartFrr()
	} else {
		fmt.Println("not start frr")
	}
}
