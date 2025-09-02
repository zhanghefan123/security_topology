package fisco_bcos_api

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"log"
	"path/filepath"
	"zhanghefan123/security_topology/api/chain_api/fisco_bcos_api/hello"
)

// CreateSessionAndDeploySmartContract 创建 session 并部署智能合约
func CreateSessionAndDeploySmartContract() *hello.HelloWorldSession {
	privateKey, _ := hex.DecodeString("145e247e170ba3afd6ae97e88f00dbc976c2345d511b0f6713355d19d8b80b58")
	//filePath := "/home/zhf/Projects/emulator/FISCO-BCOS-EXAMPLE/nodes/127.0.0.1/sdk"
	//examplePath := configs.TopConfiguration.FiscoBcosConfig.ExamplePath
	examplePath := "/home/zhf/Projects/emulator/FISCO-BCOS-EXAMPLE/"
	config := &client.Config{IsSMCrypto: false, GroupID: "group0",
		PrivateKey: privateKey, Host: "127.0.0.1", Port: 20201,
		TLSCaFile:   filepath.Join(examplePath, "nodes/127.0.0.1/sdk/ca.crt"),
		TLSKeyFile:  filepath.Join(examplePath, "nodes/127.0.0.1/sdk/sdk.key"),
		TLSCertFile: filepath.Join(examplePath, "nodes/127.0.0.1/sdk/sdk.crt"),
	}
	client, err := client.DialContext(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	input := "HelloWorld deployment 1.0"
	fmt.Println("=================DeployHelloWorld===============")
	address, receipt, instance, err := hello.DeployHelloWorld(client.GetTransactOpts(), client, input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("contract address: ", address.Hex()) // the address should be saved, will use in next example
	fmt.Println("transaction hash: ", receipt.TransactionHash)
	fmt.Println("================================")
	helloSession := &hello.HelloWorldSession{Contract: instance, CallOpts: *client.GetCallOpts(), TransactOpts: *client.GetTransactOpts()}
	version, err := helloSession.Version()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("version :", version) // "HelloWorld deployment 1.0"
	return helloSession
}
