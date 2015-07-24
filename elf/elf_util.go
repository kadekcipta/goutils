package elf

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type formatterFunc func(a byte, b ...byte) string

type field struct {
	offset byte
	name   string
	fn     formatterFunc
}

const (
	bufSize = 64
)

type HeaderInfo struct {
	Magic      string `json:"magic"`
	Class      string `json:"class"`
	Endianness string `json:"endianness"`
	Version    string `json:"version"`
	ABI        string `json:"abi"`
	ABIVersion string `json:"abi_version"`
	Type       string `json:"type"`
	Machine    string `json:"machine"`
	Entry      string `json:"entry"`
}

var fields = []field{
	field{0x00, "magic", fmtMagic},
	field{0x04, "class", func(a byte, b ...byte) string {
		if b[0] == 1 {
			return "32-Bit"
		}
		if b[0] == 2 {
			return "64-Bit"
		}
		return "Unknown"
	}},
	field{0x05, "endianness", func(a byte, b ...byte) string {
		if b[0] == 1 {
			return "Little Endian"
		}
		if b[0] == 2 {
			return "Big Endian"
		}
		return "Unknown"
	}},
	field{0x06, "version", func(a byte, b ...byte) string {
		if b[0] == 1 {
			return "1 (Original)"
		}
		return strconv.FormatInt(int64(b[0]), 10)
	}},
	field{0x07, "abi", fmtABI},
	field{0x08, "abi_version", nil},
	field{0x10, "type", fmtObjectType},
	field{0x12, "machine", fmtMachineType},
	field{0x18, "entry", fmtEntry},
}

func defaultFormatter(a byte, b ...byte) string {
	if len(b) == 0 {
		return ""
	}
	return strconv.FormatInt(int64(b[0]), 10)
}

func fmtMagic(a byte, b ...byte) string {
	if b[0] == 0x7f {
		return string(b[1:4])
	}
	return ""
}

func fmtEntry(a byte, b ...byte) string {
	numBytes := 0
	if a == 1 { // 32 bit
		numBytes = 4
	} else if a == 2 { // 64 bit
		numBytes = 8
	}

	address := uint64(0)
	if len(b) >= numBytes {
		for i := 0; i < numBytes; i++ {
			x := numBytes - i - 1
			address |= uint64(b[x]) << uint64(x*8)
		}
	}

	return fmt.Sprintf("%#x", address)
}

func fmtABI(a byte, b ...byte) string {
	switch b[0] {
	case 0x00:
		return "System V"
	case 0x01:
		return "HP-UX"
	case 0x02:
		return "NetBSD"
	case 0x03:
		return "Linux"
	case 0x06:
		return "Solaris"
	case 0x07:
		return "AIX"
	case 0x08:
		return "IRIX"
	case 0x09:
		return "FreeBSD"
	case 0x0C:
		return "OpenBSD"
	case 0x0D:
		return "OpenVMS"

	}

	return "Unknown API"
}

func fmtObjectType(a byte, b ...byte) string {
	switch b[0] {
	case 1:
		return "Relocatable"
	case 2:
		return "Executable"
	case 3:
		return "Shared"
	case 4:
		return "Core"
	}
	return "Unknown"
}

func fmtMachineType(a byte, b ...byte) string {
	switch b[0] {
	case 0x02:
		return "SPARC"
	case 0x03:
		return "x86"
	case 0x08:
		return "MIPS"
	case 0x14:
		return "PowerPC"
	case 0x28:
		return "ARM"
	case 0x2A:
		return "SuperH"
	case 0x32:
		return "IA-64"
	case 0x3E:
		return "x86-64"
	case 0xB7:
		return "AArch64"
	}
	return "Unknown"
}

func ReadHeaderInfo(path string) (*HeaderInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	var buf = [bufSize]byte{}
	n, err := reader.Read(buf[:])
	if err != nil {
		return nil, err
	}

	if n != bufSize {
		return nil, errors.New("Number of bytes not matches")
	}

	m := map[string]interface{}{}
	archCode := buf[0x04]
	for _, t := range fields {
		b := buf[t.offset:]
		fn := t.fn
		if fn == nil {
			fn = defaultFormatter
		}
		m[t.name] = fn(archCode, b...)
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	h := HeaderInfo{}
	if err := json.Unmarshal(b, &h); err != nil {
		return nil, err
	}

	if h.Magic != "ELF" {
		return nil, errors.New("Not valid ELF file")
	}

	return &h, nil
}
