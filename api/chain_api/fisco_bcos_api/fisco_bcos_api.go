package fisco_bcos_api

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"log"
	"path/filepath"
	"zhanghefan123/security_topology/api/chain_api/fisco_bcos_api/hello"
	"zhanghefan123/security_topology/configs"
)

// CreateSessionAndDeploySmartContract 创建 session 并部署智能合约
func CreateSessionAndDeploySmartContract() (*hello.HelloWorldSession, error) {
	privateKey, _ := hex.DecodeString("145e247e170ba3afd6ae97e88f00dbc976c2345d511b0f6713355d19d8b80b58")
	examplePath := configs.TopConfiguration.FiscoBcosConfig.ExamplePath
	config := &client.Config{IsSMCrypto: false, GroupID: "group0",
		PrivateKey: privateKey, Host: "127.0.0.1", Port: 20201, // 这个端口决定了应该向哪里进行转发, 20201 对应的是 FiscoBcos-2, 如果它被攻击了, 那么 TPS 将一直为0
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
		return nil, fmt.Errorf("hello world session")
	}
	fmt.Println("version :", version) // "HelloWorld deployment 1.0"
	return helloSession, nil
}
