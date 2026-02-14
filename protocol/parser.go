package protocol

import (
	"encoding/hex"
	"fmt"
)

// ResponseHeader is the expected first byte of a valid "Q" response
const ResponseHeader = 0x3d

// ParseResponse parses the raw bytes from UPS "Q" command response into UPSData.
// Response format: 3d + 7 int16 pairs (2 bytes each, big-endian) + 2 status bytes = 18 bytes
// Scale: int16 values are divided by 10
func ParseResponse(data []byte) (*UPSData, error) {
	if len(data) < 18 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	raw := data
	if len(raw) < 18 {
		return nil, fmt.Errorf("response too short: %d bytes", len(raw))
	}

	if raw[0] != ResponseHeader {
		return nil, fmt.Errorf("invalid header: expected 0x3d, got 0x%02x", raw[0])
	}

	u := &UPSData{NoData: false}

	toInt16 := func(hi, lo byte) int {
		return int(int16(hi)<<8 | int16(lo))
	}

	// Bytes 1-2: lastinputVac (big-endian int16)
	u.LastInputVac = float64(toInt16(raw[1], raw[2])) / 10
	// Bytes 3-4: inputVac
	u.InputVac = float64(toInt16(raw[3], raw[4])) / 10
	// Bytes 5-6: outputVac
	u.OutputVac = float64(toInt16(raw[5], raw[6])) / 10
	// Bytes 7-8: outputPower (percent)
	u.OutputPower = float64(toInt16(raw[7], raw[8])) / 10
	// Bytes 9-10: outputHz
	u.OutputHz = float64(toInt16(raw[9], raw[10])) / 10
	// Bytes 11-12: batterylevel
	u.BatteryLevel = float64(toInt16(raw[11], raw[12])) / 10
	// Bytes 13-14: temperatureC
	u.TemperatureC = float64(toInt16(raw[13], raw[14])) / 10

	// Bytes 15-16: status byte 1 (bits LSB=0)
	status1 := raw[15]
	u.BateriaEmUso = (status1 & (1 << 0)) != 0
	u.BateriaBaixa = (status1 & (1 << 1)) != 0
	u.ByPass = (status1 & (1 << 2)) != 0
	u.Boost = (status1 & (1 << 3)) != 0
	u.UpsOk = (status1 & (1 << 4)) != 0
	u.TesteAtivo = (status1 & (1 << 5)) != 0
	u.ShutdownAtivo = (status1 & (1 << 6)) != 0
	u.BeepLigado = (status1 & (1 << 7)) != 0

	// PowerNow computed by caller with SMSUPS_FULL_POWER
	return u, nil
}

// ParseResponseHex parses a hex string response (e.g. "3d00fa00fa...")
func ParseResponseHex(hexStr string) (*UPSData, error) {
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	return ParseResponse(raw)
}

// ApplyFullPower sets PowerNow = fullPower * (outputPower / 100)
func (u *UPSData) ApplyFullPower(fullPower int) {
	u.PowerNow = float64(fullPower) * (u.OutputPower / 100)
}
