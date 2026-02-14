package protocol

// UPSData holds parsed sensor data from the SMS Brasil UPS
type UPSData struct {
	// Voltage and power (scaled by 10 in raw response)
	LastInputVac float64
	InputVac     float64
	OutputVac    float64
	OutputPower  float64 // percent
	PowerNow     float64 // watts (SMSUPS_FULL_POWER * outputPower/100)
	OutputHz     float64

	// Battery
	BatteryLevel float64 // percent
	TemperatureC float64

	// Status flags (from status byte 1)
	BeepLigado     bool // bit 7
	ShutdownAtivo  bool // bit 6
	TesteAtivo     bool // bit 5
	UpsOk          bool // bit 4 - line OK
	Boost          bool // bit 3
	ByPass         bool // bit 2
	BateriaBaixa   bool // bit 1 - low battery
	BateriaEmUso   bool // bit 0 - battery in use

	NoData bool // true if parsing failed or no valid response
}
