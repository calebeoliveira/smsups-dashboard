# SMS UPS Dashboard

A Go desktop application that connects to SMS BRASIL UPS devices via USB serial and displays real-time sensor data in a GUI dashboard.

Compatible with the [dmslabsbr/smsUps](https://github.com/dmslabsbr/smsUps) Python project's configuration format.

<img width="609" height="566" alt="image" src="https://github.com/user-attachments/assets/b412c284-5631-49be-8c6d-08330254496d" />
Screenshot using my SMS Manager III (1500VA)

## Requirements

- Go 1.21+
- SMS BRASIL UPS connected via USB (e.g. Nobreak Manager III Senoidal)
- Linux/WSL: `/dev/ttyUSB0` or similar
- WSL Ubuntu build deps: `gcc`, `libgl1-mesa-dev`, `xorg-dev`, `libxkbcommon-dev`

## Installation

```bash
cd smsups-dashboard
go mod tidy
go build -o smsups-dashboard .
```

If you encounter checksum verification errors, try: `GOSUMDB=off go build -o smsups-dashboard .`

### WSL Ubuntu dependencies

```bash
sudo apt install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

## Configuration

Create `secrets.ini` (or `config/secrets.ini`) with the same format as the Python smsUps repo:

```ini
[config]
PORTA = /dev/ttyUSB0
INTERVALO_SERIAL = 3
BAUD_RATE = 2400

[device]
UPS_NAME = SMS
UPS_ID = 01
SMSUPS_FULL_POWER = 1400
```

- **PORTA**: Serial port path (comma-separated for fallback list)
- **INTERVALO_SERIAL**: Poll interval in seconds (default: 3)
- **BAUD_RATE**: 2400 or 9600 (Manager III often needs 2400)
- **UPS_NAME**, **UPS_ID**: Display labels
- **SMSUPS_FULL_POWER**: VA rating for power calculation (watts)

## Usage

```bash
./smsups-dashboard
# or
go run .
```

The dashboard shows:

- Input/Output voltage
- Load percent and power (watts)
- Frequency (Hz)
- Battery level and temperature
- Status: OK / On Battery / Low Battery

## Diagnostic Tool (when "Connected but no data")

Run the diagnostic to see raw bytes from the UPS:

```bash
cd smsups-dashboard
go run ./cmd/diagnose
```

It will print the raw hex response from the UPS. Share the output to debug:
- **0 bytes** = UPS not responding (wrong port, baud, or model)
- **First byte not 0x3d** = Different protocol or response format
- **Valid 3d...** = Parse issue; we can fix the parser

## WSL Notes

- USB passthrough: Attach the UPS to WSL2 (e.g. `usbipd` on Windows 11)
- Display: Use WSLg or `export DISPLAY=:0` for the Fyne GUI
- Serial device: Ensure `/dev/ttyUSB0` exists (`ls /dev/tty*`)
