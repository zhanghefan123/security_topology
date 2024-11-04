package interface_rate

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

var (
	InterfaceRateMonitorMapping = map[string]*InterfaceRateMonitor{}
)

type InterfaceRateMonitor struct {
	abstractNode      *node.AbstractNode
	TimeList          []int
	RateList          []float64
	fixedLength       int
	lastReceivedBytes int
	StopChannel       chan struct{}
}

// NewInterfaceRateMonitor 创建新的接口监听器
func NewInterfaceRateMonitor(abstractNode *node.AbstractNode) (*InterfaceRateMonitor, error) {
	interfaceRateMonitor := &InterfaceRateMonitor{
		abstractNode:      abstractNode,
		TimeList:          make([]int, 0),
		RateList:          make([]float64, 0),
		fixedLength:       10,
		lastReceivedBytes: 0,
		StopChannel:       nil, // 在启动之后会进行赋值
	}
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return nil, fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	InterfaceRateMonitorMapping[normalNode.ContainerName] = interfaceRateMonitor
	return interfaceRateMonitor, nil
}

// RemoveInterfaceRateMonitor 移除接口监听器
func RemoveInterfaceRateMonitor(abstractNode *node.AbstractNode) error {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	InterfaceRateMonitorMapping[normalNode.ContainerName].StopChannel <- struct{}{}
	delete(InterfaceRateMonitorMapping, normalNode.ContainerName)
	return nil
}

type InterfaceData struct {
	interfaceName string
	rxBytes       int
	rxPackets     int
	rxErrs        int
	rxDrops       int
	rxFifo        int
	rxFrame       int
	rxCompressed  int
	rxMulticast   int

	txBytes      int
	txPackets    int
	txErrs       int
	txDrops      int
	txFifo       int
	txFrame      int
	txCompressed int
	txMulticast  int
}

// CaptureInterfaceRate 进行接口速率的捕获
func (ir *InterfaceRateMonitor) CaptureInterfaceRate(abstractNode *node.AbstractNode) (err error) {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("GetNormalNodeFromAbstractNode: %w", err)
	}
	ir.StopChannel = ir.GetNetworkInterfaceData(normalNode)
	return nil
}

// Inter-|   Receive                                                |  Transmit
// face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
//    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
//r1_idx1:    3428      34    0    0    0     0          0         0     3584      34    0    0    0     0       0          0
//r1_idx2:    3562      35    0    0    0     0          0         0     3680      34    0    0    0     0       0          0
//  eth0:    7122      73    0    0    0     0          0         0     1026      11    0    0    0     0       0          0

func (ir *InterfaceRateMonitor) GetNetworkInterfaceData(normalNode *normal_node.NormalNode) chan struct{} {
	stopChannel := make(chan struct{})
	go func() {
		count := 0
	InternalLoop:
		for {
			select {
			case <-stopChannel:
				break InternalLoop
			default:
				content := ReadNetworkInterfaceFile(normalNode.Pid)
				networkInterfaceLines := strings.Split(content, "\n")
				firstInterfaceName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(normalNode.Type), normalNode.Id, 1)
				for _, networkInterfaceLine := range networkInterfaceLines {
					if strings.Contains(networkInterfaceLine, firstInterfaceName) {
						interfaceData := ir.ResolveNetworkInterfaceLine(networkInterfaceLine) // 第一个是 loop back
						currentReceivedBytes := interfaceData.rxBytes
						delta := float64(currentReceivedBytes - ir.lastReceivedBytes)
						dataRate := delta / float64(1024) / float64(1024)
						ir.lastReceivedBytes = currentReceivedBytes
						if len(ir.RateList) == ir.fixedLength {
							ir.RateList = ir.RateList[1:]
							ir.RateList = append(ir.RateList, dataRate)
							ir.TimeList = ir.TimeList[1:]
							ir.TimeList = append(ir.TimeList, count)
						} else {
							ir.RateList = append(ir.RateList, dataRate)
							ir.TimeList = append(ir.TimeList, count)
						}
						count += 1
						time.Sleep(time.Second)
					}
				}
			}
		}
	}()
	return stopChannel
}

// ReadNetworkInterfaceFile 进行网络接口文件的读取
func ReadNetworkInterfaceFile(pid int) string {
	var bytesContent []byte
	networkInterfaceDataFile := fmt.Sprintf("/proc/%d/net/dev", pid)
	file, err := os.Open(networkInterfaceDataFile)
	defer file.Close()
	if err != nil {
		fmt.Printf("os.Open(%s) failed: %v", networkInterfaceDataFile, err)
	}
	bytesContent, err = io.ReadAll(file)
	if err != nil {
		fmt.Printf("io.ReadAll() failed: %v", err)
	}
	stringContent := string(bytesContent)
	return stringContent
}

// ResolveNetworkInterfaceLine 进行网络接口某一行信息的解析
func (ir *InterfaceRateMonitor) ResolveNetworkInterfaceLine(networkInterfaceLine string) *InterfaceData {
	r := regexp.MustCompile("[^\\s]+")
	res := r.FindAllString(networkInterfaceLine, -1)
	rxBytes, _ := strconv.Atoi(res[1])
	rxPackets, _ := strconv.Atoi(res[2])
	rxErrs, _ := strconv.Atoi(res[3])
	rxDrops, _ := strconv.Atoi(res[4])
	rxFifo, _ := strconv.Atoi(res[5])
	rxFrame, _ := strconv.Atoi(res[6])
	rxCompressed, _ := strconv.Atoi(res[7])
	rxMulticast, _ := strconv.Atoi(res[8])
	txBytes, _ := strconv.Atoi(res[9])
	txPackets, _ := strconv.Atoi(res[10])
	txErrs, _ := strconv.Atoi(res[11])
	txDrops, _ := strconv.Atoi(res[12])
	txFifo, _ := strconv.Atoi(res[13])
	txFrame, _ := strconv.Atoi(res[14])
	txCompressed, _ := strconv.Atoi(res[15])
	txMulticast, _ := strconv.Atoi(res[16])
	interfaceData := &InterfaceData{
		interfaceName: res[0],
		rxBytes:       rxBytes,
		rxPackets:     rxPackets,
		rxErrs:        rxErrs,
		rxDrops:       rxDrops,
		rxFifo:        rxFifo,
		rxFrame:       rxFrame,
		rxCompressed:  rxCompressed,
		rxMulticast:   rxMulticast,
		txBytes:       txBytes,
		txPackets:     txPackets,
		txErrs:        txErrs,
		txDrops:       txDrops,
		txFifo:        txFifo,
		txFrame:       txFrame,
		txCompressed:  txCompressed,
		txMulticast:   txMulticast,
	}
	return interfaceData
}
