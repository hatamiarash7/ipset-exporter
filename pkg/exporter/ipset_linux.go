package exporter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
)

// IPSET_FLAG_LIST_HEADER from linux/netfilter/ipset/ip_set.h. The high-level
// netlink API always dumps members; requesting headers makes transfer and memory
// usage proportional to the number of sets rather than the number of entries.
const ipsetFlagListHeader = 1 << 2

func listIPSets() (map[string]uint32, error) {
	req := nl.NewNetlinkRequest(nl.IPSET_CMD_LIST|(unix.NFNL_SUBSYS_IPSET<<8), nl.GetIpsetFlags(nl.IPSET_CMD_LIST))
	req.AddData(&nl.Nfgenmsg{NfgenFamily: unix.AF_NETLINK, Version: nl.NFNETLINK_V0})
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_PROTOCOL, nl.Uint8Attr(nl.IPSET_PROTOCOL)))
	flags := make([]byte, 4)
	binary.BigEndian.PutUint32(flags, ipsetFlagListHeader)
	req.AddData(nl.NewRtAttr(int(nl.IPSET_ATTR_FLAGS|nl.NLA_F_NET_BYTEORDER), flags))
	counts := make(map[string]uint32)
	var parseErr error
	err := req.ExecuteIter(unix.NETLINK_NETFILTER, 0, func(msg []byte) bool {
		name, count, err := parseIPSetHeader(msg)
		if err != nil {
			parseErr = err
			return false
		}
		counts[name] = count
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list ipsets: %w", err)
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return counts, nil
}

func parseIPSetHeader(msg []byte) (string, uint32, error) {
	if len(msg) < 4 {
		return "", 0, errors.New("short ipset netfilter header")
	}
	attrs, err := parseAttributes(msg[4:])
	if err != nil {
		return "", 0, fmt.Errorf("parse ipset attributes: %w", err)
	}
	var name string
	var count uint32
	var found bool
	for _, attr := range attrs {
		switch attr.kind & nl.NLA_TYPE_MASK {
		case nl.IPSET_ATTR_SETNAME:
			name = strings.TrimRight(string(attr.value), "\x00")
		case nl.IPSET_ATTR_DATA:
			data, err := parseAttributes(attr.value)
			if err != nil {
				return "", 0, fmt.Errorf("parse ipset data: %w", err)
			}
			for _, field := range data {
				if field.kind&nl.NLA_TYPE_MASK == nl.IPSET_ATTR_ELEMENTS {
					if len(field.value) != 4 {
						return "", 0, errors.New("invalid ipset element count")
					}
					count = binary.BigEndian.Uint32(field.value)
					found = true
				}
			}
		}
	}
	if name == "" || !found {
		return "", 0, errors.New("ipset header missing set name or element count")
	}
	return name, count, nil
}

type attribute struct {
	kind  uint16
	value []byte
}

// Validate lengths including alignment padding before slicing. The dependency's
// ParseRouteAttr can panic when the final attribute has truncated padding.
func parseAttributes(data []byte) ([]attribute, error) {
	var attrs []attribute
	for len(data) > 0 {
		if len(data) < 4 {
			return nil, errors.New("short netlink attribute header")
		}
		length := int(binary.NativeEndian.Uint16(data[:2]))
		aligned := (length + 3) &^ 3
		if length < 4 || aligned > len(data) {
			return nil, errors.New("invalid netlink attribute length or padding")
		}
		attrs = append(attrs, attribute{kind: binary.NativeEndian.Uint16(data[2:4]), value: data[4:length]})
		data = data[aligned:]
	}
	return attrs, nil
}
