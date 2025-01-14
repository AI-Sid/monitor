package tools

import (
	"os"

	"github.com/getlantern/systray"
)

var appIcon []byte

var start, stop, quit *systray.MenuItem

func onStart() {
	systray.SetIcon(appIcon)
	start = systray.AddMenuItem("Start", "Start Proxy Settings Monitoring")
	stop = systray.AddMenuItem("Stop", "Stop Proxy Settings Monitoring")
	systray.AddSeparator()
	quit = systray.AddMenuItem("Quit", "Quit Proxy Settings Monitor")
    RegisterLoggingStateListener(trayLoggingModified)
    trayLoggingModified(GetLoggingEnabled())
    go handleTray()
}

func trayLoggingModified(enabled bool) {
    if enabled {
        start.Disable()
        stop.Enable()
    } else {
        start.Enable()
        stop.Disable()
    }
}

func handleTray() {
    for {
        select {
        case <- start.ClickedCh:
            SendAction(ACTION_START)
        case <- stop.ClickedCh:
            SendAction(ACTION_STOP)
        case <- quit.ClickedCh:
            SendAction(ACTION_QUIT)
            return
        }
    }
}

func onFinish() {
	// nothing
}

func RunTray() {
	if icon, err := os.ReadFile("./assets/mp.ico"); err == nil {
		appIcon = icon
	}
    RegisterQuitFunc(systray.Quit)
	systray.Run(onStart, onFinish)
}
