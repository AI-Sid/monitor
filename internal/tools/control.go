package tools

import (
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	BASE_NAME        = "Global\\AIS_Id_Proxy_Monitor"
	MUTEX_NAME       = BASE_NAME
	START_EVENT_NAME = BASE_NAME + "_Start"
	STOP_EVENT_NAME  = BASE_NAME + "_Stop"
	QUIT_EVENT_NAME  = BASE_NAME + "_Quit"
)

type Action int

const (
	ACTION_START Action = iota
	ACTION_STOP
	ACTION_QUIT
	ACTION_NONE
)

var ActionsDisplay = map[Action]string{
	ACTION_START: "START",
	ACTION_STOP: "STOP",
	ACTION_QUIT: "QUIT",
}

var actionNames = map[Action]string{
	ACTION_START: START_EVENT_NAME,
	ACTION_STOP:  STOP_EVENT_NAME,
	ACTION_QUIT:  QUIT_EVENT_NAME,
}

var internalHandles []windows.Handle = make([]windows.Handle, 0, 4)
var h2aMap, a2hMap = initMaps()

func initMaps() (map[windows.Handle]Action, map[Action]windows.Handle) {
	return make(map[windows.Handle]Action), make(map[Action]windows.Handle)
}

func clearEvents() {
	for i, h := range internalHandles {
		if i > 0 {
			windows.CloseHandle(h)
		}
	}
	internalHandles = internalHandles[:0]
	h2aMap, a2hMap = initMaps()
}

func createActionEvent(name string, action Action) error {
	if event, err := CreateNamedEvent(name); err == nil {
		internalHandles = append(internalHandles, event)
		h2aMap[event] = action
		a2hMap[action] = event
		return nil
	} else {
		return err
	}
}

func createEvents() error {
	for k, v := range actionNames {
		if err := createActionEvent(v, k); err != nil {
			return err
		}
	}
	return nil
}

func waitForActions() {
	for {
		if event, err := WaitForEvents(internalHandles...); err != nil {
			InternalError(err)
			break
		} else {
			switch h2aMap[event] {
			case ACTION_START:
				SetLoggingEnabled(true)
			case ACTION_STOP:
				SetLoggingEnabled(false)
			case ACTION_QUIT:
				HandleQuitEvent()
				return
			}
		}
	}
}

var Mutex windows.Handle = 0 // 0 indicates secondary Instance

func InitializeControl(action Action) bool {
	mtx, exists, err := CreateNamedMutex(MUTEX_NAME)
	if err != nil {
		InternalError(err)
		return false
	}
	if exists {
		return false
	}
	Mutex = mtx
	if action == ACTION_QUIT {
		return true
	}
	if err := createEvents(); err != nil {
		InternalError(err)
		return false
	}
	SetLoggingEnabled(action != ACTION_STOP)
	go waitForActions()
	return true
}

func SendAction(action Action) bool {
	if action == ACTION_NONE {
		return true
	}
	return SendNamedEvent(actionNames[action])
}

func finalizeControl() {
	clearEvents()
	if Mutex != 0 {
		syscall.CloseHandle(syscall.Handle(Mutex))
	}
}

func init() {
	RegisterFinalizer(finalizeControl)
}
