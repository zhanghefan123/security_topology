package protobuf

import "github.com/golang/protobuf/proto"

// MustMarshal marshals protobuf message to byte slice or panic when marshal twice both failed
func MustMarshal(msg proto.Message) (data []byte) {
	var err error
	defer func() {
		// while first marshal failed, retry marshal again
		if recover() != nil {
			data, err = proto.Marshal(msg)
			if err != nil {
				panic(err)
			}
		}
	}()

	data, err = proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return
}

// MustUnmarshal from byte slice to protobuf message or panic
func MustUnmarshal(b []byte, msg proto.Message) {
	if err := proto.Unmarshal(b, msg); err != nil {
		panic(err)
	}
}
