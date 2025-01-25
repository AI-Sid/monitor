package tools

import (
	//"os"
    "fmt"
	"github.com/getlantern/systray"
)

var module *ResourceModule = nil

func SetResourceModule(name string) {
    module = InitResourceModule(name)
}

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
    if module == nil {
        fmt.Println("==Current instance module creation==")
        module = InitResourceModule("")
    }
    icon, err := module.LoadIcon(100)
    if err == nil {
        appIcon, err = icon.GetIconFileBytes(-1)
        if err == nil {
            fmt.Printf("Icon loaded from resource: (size: %v)\n", len(appIcon))    
        }
    }
    if err != nil {
        fmt.Printf("Loading Icon error: %v\n", err)
    }
    RegisterQuitFunc(systray.Quit)
	systray.Run(onStart, onFinish)
}
