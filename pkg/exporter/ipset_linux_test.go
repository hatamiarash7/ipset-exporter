package exporter

import (
	"encoding/binary"
	"testing"

	"github.com/vishvananda/netlink/nl"
)

func headerMessage(name string, value []byte) []byte {
	msg := make([]byte, 4)
	msg = append(msg, nl.NewRtAttr(nl.IPSET_ATTR_SETNAME, nl.ZeroTerminated(name)).Serialize()...)
	data := nl.NewRtAttr(int(nl.IPSET_ATTR_DATA|nl.NLA_F_NESTED), nil)
	data.AddRtAttr(int(nl.IPSET_ATTR_ELEMENTS|nl.NLA_F_NET_BYTEORDER), value)
	return append(msg, data.Serialize()...)
}

func TestParseIPSetHeader(t *testing.T) {
	for _, count := range []uint32{0, 1, 1000000, 1<<32 - 1} {
		value := make([]byte, 4)
		binary.BigEndian.PutUint32(value, count)
		name, got, err := parseIPSetHeader(headerMessage("test-set", value))
		if err != nil || name != "test-set" || got != count {
			t.Fatalf("got (%q, %d, %v), want count %d", name, got, err, count)
		}
	}
	for _, msg := range [][]byte{
		nil, {0, 0, 0}, {0, 0, 0, 0},
		{0, 0, 0, 0, 255, 255, 1, 0},
		headerMessage("test", []byte{1}),
		headerMessage("", []byte{0, 0, 0, 0}),
	} {
		if _, _, err := parseIPSetHeader(msg); err == nil {
			t.Fatalf("malformed message accepted: %v", msg)
		}
	}
}

func FuzzParseIPSetHeader(f *testing.F) {
	f.Add(headerMessage("test", []byte{0, 0, 0, 1}))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _, _ = parseIPSetHeader(data)
	})
}
