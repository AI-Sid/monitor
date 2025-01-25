package tools

import (
	"fmt"
	"sync"
    "unsafe"
	"golang.org/x/sys/windows"
)

type Handle = windows.Handle

type ResourceID struct {
    base windows.ResourceIDOrString
    stored *uint16
    ptr uintptr
}

func checkInt(value int64) windows.ResourceIDOrString {
    if value > 0 && value < 0x8000 {
        return windows.ResourceID(value)
    }
    fmt.Printf("Int checking error: %v\n", value)
    return windows.ResourceID(0)
}

func checkUint(value uint64) windows.ResourceIDOrString {
    if value > 0 && value < 0x8000 {
        return windows.ResourceID(value)
    }
    fmt.Printf("Uint checking error: %v\n", value)
    return windows.ResourceID(0)
}

func normalizeResourceIDOrString(value windows.ResourceIDOrString) windows.ResourceIDOrString {
    switch v :=value.(type) {
    case windows.ResourceID:
        return value
    case string:
        return value
    case int: 
        return checkInt(int64(v))
    case int8:
        return checkInt(int64(v))
    case int16:
        return checkInt(int64(v))
    case int32:
        return checkInt(int64(v))
    case int64:
        return checkInt(v)
    case byte:
        return checkUint(uint64(v))
    case uint16:
        return checkUint(uint64(v))
    case uint32:
        return checkUint(uint64(v))
    case uint64:
        return checkUint(uint64(v))
    default:
        fmt.Printf("!!!Errorneous resource id: %v\n", v)    
    }
    fmt.Printf("Errorneous resource id: %v\n", value)
    return windows.ResourceID(0)
}

func MakeIntResource(name windows.ResourceIDOrString) (*ResourceID, error) {
    name = normalizeResourceIDOrString(name)
    res := new(ResourceID)
    res.base = name
    var err error
    switch v := name.(type) {
    case string:
        res.stored, err = GetUint16String(v)
        if err != nil {
            return nil, err
        }
        res.ptr = uintptr(unsafe.Pointer(res.stored))
        return res, nil
    case windows.ResourceID:
        res.ptr = uintptr(v)
        return res, nil
    }
    return nil, fmt.Errorf("parameter must be a ResourceID or a string")
}

func (v *ResourceID) GetPtr() uintptr {
    if v == nil {
        return 0
    }
    return v.ptr
}

func (v *ResourceID) GetValue() windows.ResourceIDOrString {
    return v.base
}

func (v *ResourceID) Close() {
    if v != nil {
        v.base = nil
        v.stored = nil
        v.ptr = 0
    }
}

type ResourceModule struct {
    name string
    isValid bool
    dllImported bool
    handle Handle
    createErr error
    moduleErr error
}

type ResourceInfo struct {
    module *ResourceModule
    info Handle
}

func newResourceInfo(module *ResourceModule, info Handle) *ResourceInfo {
    if module == nil || info == 0 {
        return nil
    }
    return &ResourceInfo{module, info}
}

func (v *ResourceInfo) GetHandles() (moduleHandle, infoHandle Handle) {
    return v.module.handle, v.info
}

func (v *ResourceInfo) LoadData() ([]byte, error) {
    return windows.LoadResourceData(v.module.handle, v.info)
}

var instanceResourceModule *ResourceModule
var otherModules = map[string]*ResourceModule{}
var lock sync.Mutex

var errInstanceModuleCreation = fmt.Errorf("instance module is invalid")
var customModuleErrorString = "Module %v is invalid"

func ResouceDLLFileName(name string) string {
    return "debug/"+name+"Res.dll"
}

func InitResourceModule(name string) *ResourceModule {
    lock.Lock()
    defer lock.Unlock()
    var handle Handle
    var err error
    if len(name) == 0 {
        if instanceResourceModule != nil {
            return instanceResourceModule
        }
        err = windows.GetModuleHandleEx(0, nil, &handle)
        if err == nil {
            instanceResourceModule = &ResourceModule{name: "", isValid: true, dllImported: false, handle: handle}
            return instanceResourceModule
        } else {
            fmt.Printf("Instance module error: %v\n", err)
            instanceResourceModule = &ResourceModule{name: "", isValid: false, dllImported: false, handle: 0, createErr: err, moduleErr: errInstanceModuleCreation}
        }
        return instanceResourceModule           
    } else {
        if m, ok := otherModules[name]; ok {
            return m
        }
        var res *ResourceModule
        dllFileName := ResouceDLLFileName(name)
        handle, err = windows.LoadLibrary(dllFileName)
        if err == nil {
            fmt.Printf("Dll file %v found\n", dllFileName)
            res = &ResourceModule{name:name, isValid:true, dllImported:true, handle:handle}
        } else {
            fmt.Printf("Dll file debug/%vRes.dll open error", name)
            res = &ResourceModule{name:name, isValid:false, dllImported: false, handle: 0, createErr: err, moduleErr: fmt.Errorf(customModuleErrorString, name)}
        }
        otherModules[name]=res
        return res
    }
}

func (v *ResourceModule) IsValid() bool {
    return v.isValid
}

func (v *ResourceModule) IsMainInstance() bool {
    return v.isValid && !v.dllImported
}

func (v *ResourceModule) GetName() string {
    return v.name
}

func (v *ResourceModule) GetResourceInfo(name windows.ResourceIDOrString, resourceType windows.ResourceIDOrString) (*ResourceInfo, error) {
    if !v.isValid {
        return nil, v.moduleErr
    }
    name = normalizeResourceIDOrString(name)
    if v,ok := name.(windows.ResourceID); ok {
        fmt.Printf("Resource ID: %v\n", v)    
    } else {
        fmt.Printf("Resource Name: %v\n", name)    
    }
    if info, err := windows.FindResource(v.handle, name, resourceType); err != nil {
        return nil, err
    } else {
        return newResourceInfo(v, info), nil
    }
}

func (v *ResourceModule) LoadIcon(name windows.ResourceIDOrString) (*Icon, error) {
    if !v.isValid {
        return nil, v.moduleErr
    }
    rid, err := MakeIntResource(name)
    if err != nil {
        return nil, err
    }
    return newIcon(v, rid)
}

func releaseModules() {
    for _, m := range otherModules {
        if m.isValid && m.dllImported {
            windows.FreeLibrary(m.handle)
        }
    }
    otherModules = make(map[string]*ResourceModule)
    if instanceResourceModule != nil && instanceResourceModule.isValid {
        windows.CloseHandle(instanceResourceModule.handle)
        instanceResourceModule = nil
    }
}

func init() {
    RegisterQuitFunc(releaseModules)
}
