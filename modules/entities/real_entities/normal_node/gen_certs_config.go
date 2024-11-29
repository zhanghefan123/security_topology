package normal_node

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
)

// GeneratePeerIdAndPrivateKey 进行密钥的生成
func (normalNode *NormalNode) GeneratePeerIdAndPrivateKey() (err error) {
	var peerIdPath, privateKeyPath string
	err, peerIdPath, privateKeyPath = normalNode.GetPeerIdAndPrivateKeyPath()
	if err != nil {
		return fmt.Errorf("GetPeerIdAndPrivateKeyPath err: %w", err)
	}

	var peerIdFile *os.File
	var privateKeyFile *os.File
	peerIdFile, err = os.Create(peerIdPath)
	defer func() {
		closeFileErr := peerIdFile.Close()
		if err == nil {
			err = closeFileErr
		}
	}()
	if err != nil {
		return fmt.Errorf("generate peer id file err: %w", err)
	}

	privateKeyFile, err = os.Create(privateKeyPath)
	defer func() {
		fileCloseErr := privateKeyFile.Close()
		if err == nil {
			err = fileCloseErr
		}
	}()
	if err != nil {
		return fmt.Errorf("generate private key file err: %w", err)
	}

	// 产生私钥匙
	var privateKey crypto.PrivKey
	privateKey, _, err = crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return fmt.Errorf("generate peer id and private key error: %w", err)
	}

	// 将私钥编码为 pem 格式
	var privateBytes []byte
	privateBytes, err = crypto.MarshalPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("generate peer id and private key error: %w", err)
	}

	// 将私钥写入文件
	_, err = privateKeyFile.Write(privateBytes)
	if err != nil {
		return fmt.Errorf("generate peer id and private key error: %w", err)
	}

	// 创建 peerId
	peerID, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("generate peer id from private key error: %w", err)
	}

	// 写入 peerId
	_, err = peerIdFile.Write([]byte(peerID.String()))
	if err != nil {
		return fmt.Errorf("error writing peer ID: %v", err)
	}
	return nil
}

// GetPeerIdAndPrivateKeyPath 获取 peerid 和 privatekey
func (normalNode *NormalNode) GetPeerIdAndPrivateKeyPath() (error, string, string) {
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	outputDir := filepath.Join(simulationDir, normalNode.ContainerName, "security")

	// 创建目录
	err := dir.Generate(outputDir)
	if err != nil {
		return fmt.Errorf("get peer id and private key path failed: %w", err), "", ""
	}

	peerIdPath := filepath.Join(outputDir, "peerId")
	privateKeyPath := filepath.Join(outputDir, "private.key")
	return nil, peerIdPath, privateKeyPath
}
