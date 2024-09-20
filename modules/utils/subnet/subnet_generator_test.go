package subnet

import "testing"

func TestSubnetGenerator(t *testing.T) {
	GenerateSubnets("192.168.0.0/16")
}
