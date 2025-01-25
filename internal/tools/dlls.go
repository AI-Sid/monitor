package tools

import (
    "syscall"
    //"golang.org/x/sys/windows"
)

type LazyDLL struct {
    lazyDll *syscall.LazyDLL
    funcs map[string]*syscall.LazyProc
}

func (v *LazyDLL) GetProc(name string) *syscall.LazyProc {
    if v == nil {
        return nil
    }
    if res, ok := v.funcs[name]; ok {
        return res
    }
    res := v.lazyDll.NewProc(name)
    v.funcs[name] = res
    return res
}

var dlls = make(map[string]*LazyDLL)

func GetDll(name string) *LazyDLL {
    res := dlls[name]
    if res == nil {
        res = &LazyDLL{syscall.NewLazyDLL(name), make(map[string]*syscall.LazyProc)}
        dlls[name] = res
    }
    return res
}

func GetDllProc(dllName, procName string) *syscall.LazyProc {
    dll := GetDll(dllName)
    if dll == nil {
        return nil
    }
    return dll.GetProc(procName)
}



