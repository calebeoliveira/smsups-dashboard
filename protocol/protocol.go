package protocol

// QueryCommand is the hex for "Q" command: 51 ff ff ff ff b3 0d
// Get UPS data (primary polling command)
var QueryCommand = []byte{0x51, 0xff, 0xff, 0xff, 0xff, 0xb3, 0x0d}

// BuildQueryCommand returns the "Q" command bytes for polling UPS data.
// Byte-for-byte match with Python: "51 ff ff ff ff b3 0d"
func BuildQueryCommand() []byte {
	cmd := make([]byte, len(QueryCommand))
	copy(cmd, QueryCommand)
	return cmd
}

// Checksum computes 8-bit checksum: (0x100 - (sum % 0x100)) % 0x100
// Used when building custom commands. The "Q" command is pre-computed.
func Checksum(data []byte) byte {
	var sum int
	for _, b := range data {
		sum += int(b)
	}
	sum = sum % 0x100
	ret := 0x100 - sum
	if ret == 256 {
		ret = 0
	}
	return byte(ret)
}
