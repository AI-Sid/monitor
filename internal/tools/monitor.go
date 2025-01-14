package tools

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	INET_KEY   = `SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`
	REG_NOTIFY = 5 // REG_NOTIFY_CHANGE_NAME | REG_NOTIFY_CHANGE_LAST_SET
)

var (
	advapi32                   = syscall.NewLazyDLL("Advapi32.dll")
	winRegNotifyChangeKeyValue = advapi32.NewProc("RegNotifyChangeKeyValue")
)

var cancel windows.Handle
var loggingEnabled bool

type MonitorStateChanged = func(value bool)

var changeStateLocked bool = false
var listeners []MonitorStateChanged = make([]MonitorStateChanged, 0, 4)
var monitorMutex sync.Mutex

var proxyEnabled bool
var proxyServer string

func RegisterLoggingStateListener(listener MonitorStateChanged) {
	if listener == nil {
		return
	}
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	listeners = append(listeners, listener)
}

const timeFormat = "2006-01-02T15:04:05.000"

func logProxyData(log *os.File) {
	timestamp := time.Now().Format(timeFormat)
	if proxyEnabled {
		fmt.Fprintf(log, "%v        proxy on, %v\n", timestamp, proxyServer)
	} else {
		fmt.Fprintf(log, "%v        proxy off\n", timestamp)
	}
}

func SetLoggingEnabled(value bool) {
    if cancel == 0 {
        return
    }
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
    if loggingEnabled == value {
		return
	}
    if changeStateLocked {
		fmt.Println("Error: Monitor state is locked.")
        return
	}
	changeStateLocked = true
    defer func() {
        changeStateLocked = false
    }()
	if value {
		err := startMonitor()
		if err != nil {
			InternalError(err)
			return
		}
	} else {
		if cancel != 0 {
			err := windows.SetEvent(cancel)
			if err != nil {
				InternalError(err)
				return
			}
		}
	}
	loggingEnabled = value
	for _, listener := range listeners {
		func() {
			monitorMutex.Unlock()
			defer monitorMutex.Lock()
			listener(value)
		}()
	}
}

func GetLoggingEnabled() bool {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	return loggingEnabled
}

type monitorState struct {
    monitoring bool
    log *os.File
    key registry.Key
    event windows.Handle
}

func (v *monitorState) Release(force bool) {
    if !force && v.monitoring {
        return
    }
    if v.key != 0 {
		v.key.Close()
        v.key = 0
	}
	if v.event != 0 {
		CloseEvent(&v.event)
	}
	if v.log != nil {
		v.log.Close()
        v.log = nil
	}
}

func startMonitor() error {
	var err error
    state := &monitorState{}
	defer state.Release(false)
	filePath := os.Getenv("APPDATA") + "/appname"
	if err := os.MkdirAll(filePath, 00770); err != nil {
		return err
	}
	filePath += "/appname.log"
	state.log, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if state.key, err = registry.OpenKey(registry.CURRENT_USER, INET_KEY, registry.NOTIFY); err != nil {
		return err
	}
	state.event, err = CreateEvent()
	if err != nil {
		return err
	}
	state.monitoring = true
	go monitoring(state)
	return nil
}

func monitoring(state *monitorState) {
	defer state.Release(true)
	firstCall := true
	for GetLoggingEnabled() {
		ret, _, err := winRegNotifyChangeKeyValue.Call(uintptr(state.key), 0, REG_NOTIFY, uintptr(state.event), 1)
		if ret != uintptr(windows.ERROR_SUCCESS) {
			InternalError(err)
			break
		}
		err = updateProxySettings(&firstCall, state.log)
        if err != nil {
            InternalError(err)
            break
        }
		event, err := WaitForEvents(cancel, state.event)
		if err != nil {
			InternalError(err)
			break
		}
		if event == cancel {
			break
		}
	}
}

func updateProxySettings(firstCall *bool, log *os.File) error {
	var (
		enabled    bool
		server     string
		err        error
		enabledInt uint64
	)
	var k registry.Key
	k, err = registry.OpenKey(registry.CURRENT_USER, INET_KEY, registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	enabledInt, _, err = k.GetIntegerValue("ProxyEnable")
	if err != nil {
        enabledInt = 0
	}
	enabled = enabledInt != 0
	server, _, err = k.GetStringValue("ProxyServer")
	if err != nil {
		server = ""
	}
	if *firstCall || proxyEnabled != enabled || proxyServer != server {
		proxyEnabled, proxyServer, *firstCall = enabled, server, false
		logProxyData(log)
	}
	return nil
}

func finalizeMonitor() {
    SetLoggingEnabled(false)
    if cancel != 0 {
        CloseEvent(&cancel)
    }
}

func init() {
    RegisterFinalizer(finalizeMonitor)
    c, err := CreateEvent()
    if err != nil {
        InternalError(err)
    } else {
        cancel = c
    }
}
