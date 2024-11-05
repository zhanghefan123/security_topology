package interface_data

import (
	"regexp"
	"strconv"
)

type InterfaceData struct {
	interfaceName string
	RxBytes       int
	RxPackets     int
	RxErrs        int
	RxDrops       int
	RxFifo        int
	RxFrame       int
	RxCompressed  int
	RxMulticast   int

	TxBytes      int
	TxPackets    int
	TxErrs       int
	TxDrops      int
	TxFifo       int
	TxFrame      int
	TxCompressed int
	TxMulticast  int
}

// ResolveNetworkInterfaceLine 进行网络接口某一行信息的解析 -> 并且返回数据
func ResolveNetworkInterfaceLine(networkInterfaceLine string) *InterfaceData {
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
		RxBytes:       rxBytes,
		RxPackets:     rxPackets,
		RxErrs:        rxErrs,
		RxDrops:       rxDrops,
		RxFifo:        rxFifo,
		RxFrame:       rxFrame,
		RxCompressed:  rxCompressed,
		RxMulticast:   rxMulticast,
		TxBytes:       txBytes,
		TxPackets:     txPackets,
		TxErrs:        txErrs,
		TxDrops:       txDrops,
		TxFifo:        txFifo,
		TxFrame:       txFrame,
		TxCompressed:  txCompressed,
		TxMulticast:   txMulticast,
	}
	return interfaceData
}
