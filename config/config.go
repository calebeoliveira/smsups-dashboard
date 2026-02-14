package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Config holds application configuration compatible with smsUps secrets.ini
type Config struct {
	// From [config]
	PORTA          []string // Serial port(s), tried in order
	IntervalSerial int      // Poll interval in seconds
	BaudRate       int      // Serial baud rate (2400 or 9600 - Manager III varies by unit)

	// From [device]
	UPSName        string
	UPSID          string
	SMSUPSFullPower int
}

// Default returns a Config with sensible defaults
func Default() *Config {
	return &Config{
		PORTA:           []string{"/dev/ttyUSB0"},
		IntervalSerial:   3,
		BaudRate:         2400, // Manager III often uses 2400 (some use 9600)
		UPSName:          "SMS",
		UPSID:            "01",
		SMSUPSFullPower:  1400,
	}
}

// Load reads configuration from secrets.ini or config.json (Home Assistant options schema)
func Load() (*Config, error) {
	cfg := Default()

	// Try secrets.ini first
	if p := findConfigFile("secrets.ini"); p != "" {
		if err := loadINI(cfg, p); err != nil {
			return cfg, err
		}
	}

	// Try config.json (HA add-on options)
	if p := findConfigFile("config.json"); p != "" {
		if err := loadJSON(cfg, p); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

func findConfigFile(name string) string {
	dirs := []string{".", "config", "./config", "smsups-dashboard", "smsups-dashboard/config"}
	for _, d := range dirs {
		p := filepath.Join(d, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func loadINI(cfg *Config, path string) error {
	iniCfg, err := ini.Load(path)
	if err != nil {
		return err
	}

	if s := iniCfg.Section("config"); s != nil {
		if v := s.Key("PORTA").String(); v != "" {
			cfg.PORTA = splitPorts(v)
		}
		if v, err := s.Key("INTERVALO_SERIAL").Int(); err == nil && v > 0 {
			cfg.IntervalSerial = v
		}
		if v, err := s.Key("BAUD_RATE").Int(); err == nil && (v == 2400 || v == 4800 || v == 9600) {
			cfg.BaudRate = v
		}
	}

	if s := iniCfg.Section("device"); s != nil {
		if v := s.Key("UPS_NAME").String(); v != "" {
			cfg.UPSName = strings.TrimSpace(v)
		}
		if v := s.Key("UPS_ID").String(); v != "" {
			cfg.UPSID = strings.TrimSpace(v)
		}
		if v, err := s.Key("SMSUPS_FULL_POWER").Int(); err == nil && v > 0 {
			cfg.SMSUPSFullPower = v
		}
	}

	return nil
}

func splitPorts(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(strings.TrimSpace(p), `"'`))
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"/dev/ttyUSB0"}
	}
	return out
}

// haOptions matches Home Assistant add-on options schema
type haOptions struct {
	Options struct {
		PORTA             string `json:"PORTA"`
		BAUD_RATE         int    `json:"BAUD_RATE"`
		SMSUPS_FULL_POWER int    `json:"SMSUPS_FULL_POWER"`
		UPS_NAME          string `json:"UPS_NAME"`
		UPS_ID            string `json:"UPS_ID"`
	} `json:"options"`
}

func loadJSON(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var opts haOptions
	if err := json.Unmarshal(data, &opts); err != nil {
		return err
	}

	o := opts.Options
	if o.PORTA != "" {
		cfg.PORTA = splitPorts(o.PORTA)
	}
	if o.BAUD_RATE == 2400 || o.BAUD_RATE == 4800 || o.BAUD_RATE == 9600 {
		cfg.BaudRate = o.BAUD_RATE
	}
	if o.SMSUPS_FULL_POWER > 0 {
		cfg.SMSUPSFullPower = o.SMSUPS_FULL_POWER
	}
	if o.UPS_NAME != "" {
		cfg.UPSName = o.UPS_NAME
	}
	if o.UPS_ID != "" {
		cfg.UPSID = o.UPS_ID
	}

	return nil
}
