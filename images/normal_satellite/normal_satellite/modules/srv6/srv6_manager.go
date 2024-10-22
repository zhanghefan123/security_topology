package srv6

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ReadSRv6Routes 进行路由的读取
func ReadSRv6Routes() (result []string) {
	containerName := os.Getenv("CONTAINER_NAME")

	srv6FilePath := fmt.Sprintf("/configuration/%s/route/srv6.txt", containerName)

	var err error
	var file *os.File
	var all []byte

	file, err = os.Open(srv6FilePath)
	if err != nil {
		return nil
	}

	all, err = io.ReadAll(file)
	if err != nil {
		return nil
	}

	srv6Routes := strings.Split(string(all), "\n")

	return srv6Routes
}

// InsertSRv6Routes 进行路由的插入
func InsertSRv6Routes(srv6Routes []string) {
	startInsertSignal := make(chan struct{})

	// 这里需要进行睡眠5s的原因是当接口没有完全建立起来的时候, 进行路由的插入, 会造成路由插入的失败
	go func() {
		time.Sleep(5 * time.Second)
		close(startInsertSignal)
	}()

	<-startInsertSignal

	for _, route := range srv6Routes {
		routeSplit := strings.Split(route, " ")
		cmd := exec.Command(routeSplit[0], routeSplit[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Println(string(output))
	}
}

func Start() {
	srv6Routes := ReadSRv6Routes()
	InsertSRv6Routes(srv6Routes)
}
