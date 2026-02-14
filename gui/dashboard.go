package gui

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"smsups-dashboard/config"
	"smsups-dashboard/protocol"
	"smsups-dashboard/serial"
)

// Status colors
var (
	colorOK     = color.NRGBA{R: 46, G: 204, B: 113, A: 255}  // green
	colorBattery = color.NRGBA{R: 241, G: 196, B: 15, A: 255}  // amber
	colorLow    = color.NRGBA{R: 231, G: 76, B: 60, A: 255}   // red
	colorOff    = color.NRGBA{R: 149, G: 165, B: 166, A: 255} // gray
)

// Dashboard runs the Fyne GUI and polling loop
type Dashboard struct {
	app    fyne.App
	window fyne.Window

	// Display widgets
	connLabel   *widget.Label
	statusText  *canvas.Text
	batteryBar  *widget.ProgressBar
	loadBar       *widget.ProgressBar
	valueLabels   map[string]*widget.Label
	debugLabel    *widget.Label

	// State
	latestData  *protocol.UPSData
	latestMutex sync.RWMutex
	connected   bool
}

// New creates and configures the dashboard
func New(cfg *config.Config) *Dashboard {
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("SMS UPS - %s_%s", cfg.UPSName, cfg.UPSID))
	w.Resize(fyne.NewSize(580, 420))

	d := &Dashboard{
		app:         a,
		window:      w,
		valueLabels: make(map[string]*widget.Label),
	}

	d.buildUI(cfg)
	return d
}

func (d *Dashboard) buildUI(cfg *config.Config) {
	d.connLabel = widget.NewLabelWithStyle("Connecting...", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	d.statusText = canvas.NewText("--", colorOff)
	d.statusText.TextSize = 20
	d.statusText.TextStyle = fyne.TextStyle{Bold: true}
	d.batteryBar = widget.NewProgressBar()
	d.batteryBar.Min = 0
	d.batteryBar.Max = 1
	d.loadBar = widget.NewProgressBar()
	d.loadBar.Min = 0
	d.loadBar.Max = 1
	d.debugLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})

	// Header: connection left, colored status center
	header := container.NewBorder(nil, nil, d.connLabel, nil,
		container.NewCenter(container.NewPadded(d.statusText)),
	)

	// Uniform card: centered title, icon, value, optional bar. All elements center-aligned.
	makeCard := func(title string, icon fyne.Resource, valueKey string) fyne.CanvasObject {
		l := widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		d.valueLabels[valueKey] = l
		titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{})
		var content *fyne.Container
		if icon != nil {
			content = container.NewVBox(
				container.NewCenter(titleLbl),
				container.NewCenter(widget.NewIcon(icon)),
				container.NewCenter(l),
			)
		} else {
			content = container.NewVBox(
				container.NewCenter(titleLbl),
				container.NewCenter(l),
			)
		}
		return widget.NewCard("", "", content)
	}

	makeCardWithBar := func(title string, valueKey string, bar *widget.ProgressBar, extraKeys ...string) fyne.CanvasObject {
		l := widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		d.valueLabels[valueKey] = l
		titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{})
		icon := theme.StorageIcon()
		if valueKey == "power" {
			icon = theme.WarningIcon()
		}
		row := container.NewVBox(
			container.NewCenter(titleLbl),
			container.NewCenter(widget.NewIcon(icon)),
			container.NewCenter(l),
		)
		for _, k := range extraKeys {
			el := widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{})
			d.valueLabels[k] = el
			row.Add(container.NewCenter(el))
		}
		row.Add(bar)
		return widget.NewCard("", "", row)
	}

	inputCard := makeCard("Input", theme.HomeIcon(), "input")
	outputCard := makeCard("Output", theme.StorageIcon(), "output")
	freqCard := makeCard("Frequency", theme.ViewFullScreenIcon(), "hz")
	loadCard := makeCardWithBar("Load", "power", d.loadBar, "watts")
	batteryCard := makeCardWithBar("Battery", "battery", d.batteryBar)
	tempCard := makeCard("Temperature", theme.SettingsIcon(), "temp")

	metricsGrid := container.NewGridWithColumns(3,
		container.NewPadded(inputCard),
		container.NewPadded(outputCard),
		container.NewPadded(freqCard),
		container.NewPadded(loadCard),
		container.NewPadded(batteryCard),
		container.NewPadded(tempCard),
	)

	content := container.NewVBox(
		container.NewPadded(header),
		widget.NewSeparator(),
		container.NewPadded(metricsGrid),
		container.NewPadded(d.debugLabel),
	)

	d.window.SetContent(content)
}

// statusColor returns color for status
func statusColor(u *protocol.UPSData, connected bool) color.Color {
	if !connected || u == nil || u.NoData {
		return colorOff
	}
	if u.BateriaBaixa {
		return colorLow
	}
	if u.BateriaEmUso {
		return colorBattery
	}
	return colorOK
}

// setDataWithDebug updates the displayed UPS data and optional debug info
func (d *Dashboard) setDataWithDebug(u *protocol.UPSData, connected bool, debugMsg string) {
	d.latestMutex.Lock()
	d.latestData = u
	d.connected = connected
	d.latestMutex.Unlock()

	if connected {
		d.connLabel.SetText("● Connected")
	} else {
		d.connLabel.SetText("○ Disconnected")
	}

	if u != nil && !u.NoData {
		// Status with color
		status := StatusText(u)
		d.statusText.Text = status
		d.statusText.Color = statusColor(u, connected)
		d.statusText.Refresh()

		// Values
		d.valueLabels["input"].SetText(fmt.Sprintf("%.1f V", u.InputVac))
		d.valueLabels["output"].SetText(fmt.Sprintf("%.1f V", u.OutputVac))
		d.valueLabels["power"].SetText(fmt.Sprintf("%.0f%%", u.OutputPower))
		if wl, ok := d.valueLabels["watts"]; ok {
			wl.SetText(fmt.Sprintf("%.0f W", u.PowerNow))
		}
		d.valueLabels["hz"].SetText(fmt.Sprintf("%.1f Hz", u.OutputHz))
		d.valueLabels["battery"].SetText(fmt.Sprintf("%.0f%%", u.BatteryLevel))
		d.valueLabels["temp"].SetText(fmt.Sprintf("%.0f °C", u.TemperatureC))

		// Progress bars (0-1)
		d.batteryBar.SetValue(u.BatteryLevel / 100)
		d.loadBar.SetValue(u.OutputPower / 100)
	} else {
		d.statusText.Text = "No data"
		d.statusText.Color = colorOff
		d.statusText.Refresh()
		for _, l := range d.valueLabels {
			l.SetText("--")
		}
		d.batteryBar.SetValue(0)
		d.loadBar.SetValue(0)
	}

	if debugMsg != "" {
		d.debugLabel.SetText("Debug: " + debugMsg)
		d.debugLabel.Show()
	} else {
		d.debugLabel.Hide()
	}
}

// Run starts the UI and polling loop
func (d *Dashboard) Run(cfg *config.Config) {
	var port *serial.Port
	queryCmd := protocol.BuildQueryCommand()

	go func() {
		for {
			if port == nil {
				p, err := serial.OpenWithBaud(cfg.PORTA, cfg.BaudRate)
				if err != nil {
					d.setDataWithDebug(nil, false, "Open: "+err.Error())
				} else {
					port = p
					port.Wake()
					d.setDataWithDebug(nil, true, "")
				}
			}

			if port != nil {
				resp, err := port.SendCommand(queryCmd)
				if err != nil {
					port.Close()
					port = nil
					d.setDataWithDebug(nil, false, "Read: "+err.Error())
				} else {
					u, err := protocol.ParseResponse(resp)
					if err != nil {
						rawHex := hex.EncodeToString(resp)
						if len(rawHex) > 60 {
							rawHex = rawHex[:60] + "..."
						}
						d.setDataWithDebug(&protocol.UPSData{NoData: true}, true, fmt.Sprintf("%d bytes [%s] - %v", len(resp), rawHex, err))
					} else {
						u.ApplyFullPower(cfg.SMSUPSFullPower)
						d.setDataWithDebug(u, true, "")
					}
				}
			}

			interval := time.Duration(cfg.IntervalSerial) * time.Second
			if interval < time.Second {
				interval = time.Second
			}
			time.Sleep(interval)
		}
	}()

	d.window.ShowAndRun()
}
