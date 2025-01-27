package tools

import (
	"fmt"
	"strconv"
	"sync"
	"syscall"

	"golang.org/x/sys/windows"
)

var stateIsNormal = true
var stateMutex sync.Mutex
var buildMode = false

func SetBuildMode(value string) bool {
	buildMode, _ = strconv.ParseBool(value)
	return buildMode
}

func GetBuildMode() bool {
	return buildMode
}

func InternalError(err error) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	stateIsNormal = false
	fmt.Printf("Internal Error: %v\n", err)
}

func IsNormalState() bool {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	return stateIsNormal
}

func PanicHandler(isError *bool) {
	if err := recover(); err != nil {
		if isError != nil {
			*isError = true
		}
		InternalError(err.(error))
	}
}

type SimpleFunc = func()

var finalizeFuncs = make([]SimpleFunc, 0, 4)

func RegisterFinalizer(f SimpleFunc) {
	finalizeFuncs = append(finalizeFuncs, f)
}

func callSimpleFunc(f SimpleFunc) {
	defer PanicHandler(nil)
	f()
}

var quitFuncs []SimpleFunc = make([]SimpleFunc, 0, 4)

func RegisterQuitFunc(f SimpleFunc) {
	quitFuncs = append(quitFuncs, f)
}

func HandleQuitEvent() {
	for i := len(quitFuncs) - 1; i >= 0; i-- {
		callSimpleFunc(quitFuncs[i])
	}
}

func DoExitProgram() {
	for i := len(finalizeFuncs) - 1; i >= 0; i-- {
		callSimpleFunc(finalizeFuncs[i])
	}
}

func GetUint16String(v string) (*uint16, error) {
	return syscall.UTF16PtrFromString(v)
}

func CreateNamedMutex(name string) (handle windows.Handle, exists bool, e error) {
	n, err := GetUint16String(name)
	if err != nil {
		return 0, false, err
	}
	handle, err = windows.CreateMutex(nil, false, n)
	if err != nil {
		if err.(syscall.Errno) == syscall.ERROR_ALREADY_EXISTS {
			return 0, true, nil
		}
		return 0, false, err
	}
	return handle, false, nil
}

func CreateNamedEvent(name string) (windows.Handle, error) {
	var nptr *uint16
	if name != "" {
		if n, err := GetUint16String(name); err != nil {
			return 0, err
		} else {
			nptr = n
		}
	}
	e, err := windows.CreateEvent(nil, 0, 0, nptr)
	if err != nil {
		return 0, err
	}
	return e, nil
}

func CreateEvent() (windows.Handle, error) {
	return CreateNamedEvent("")
}

func SendNamedEvent(name string) bool {
	var e error
	if nptr, err := GetUint16String(name); err == nil {
		if event, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, nptr); err == nil {
			defer syscall.CloseHandle(syscall.Handle(event))
			e = windows.SetEvent(event)
		} else {
			e = err
		}
	} else {
		e = err
	}
	if e != nil {
		InternalError(e)
	}
	return e == nil
}

func WaitForEvents(events ...windows.Handle) (windows.Handle, error) {
	idx, err := windows.WaitForMultipleObjects(events, false, windows.INFINITE)
	if err != nil {
		return 0, err
	}
	idx -= windows.WAIT_OBJECT_0 // formal, because WAIT_OBJECT_0 == 0
	return events[idx], nil
}

func CloseEvent(value *windows.Handle) {
	if value == nil || *value == 0 {
		return
	}
	err := syscall.CloseHandle(syscall.Handle(*value))
    if err != nil {
        InternalError(err)
    }
	*value = 0
}
