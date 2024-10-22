package frr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

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

func StartFrr() {
	cmd := exec.Command("service", "frr", "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}

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
