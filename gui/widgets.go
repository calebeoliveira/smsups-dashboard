package gui

import (
	"fmt"

	"smsups-dashboard/protocol"
)

// FormatUPSData formats UPSData for display
func FormatUPSData(u *protocol.UPSData) string {
	if u == nil || u.NoData {
		return "No data"
	}
	return fmt.Sprintf("In: %.1fV | Out: %.1fV (%.0f%%) | %.0fW | %.1fHz | Bat: %.0f%% | %.0f°C",
		u.InputVac, u.OutputVac, u.OutputPower, u.PowerNow, u.OutputHz, u.BatteryLevel, u.TemperatureC)
}

// StatusText returns human-readable status
func StatusText(u *protocol.UPSData) string {
	if u == nil || u.NoData {
		return "Disconnected"
	}
	if u.BateriaBaixa {
		return "Low Battery"
	}
	if u.BateriaEmUso {
		return "On Battery"
	}
	if u.UpsOk {
		return "OK"
	}
	return "Unknown"
}
