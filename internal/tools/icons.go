package tools

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/windows"
	//"golang.org/x/sys/windows"
)

var (
    loadIcon = GetDllProc("user32.dll", "LoadIconW")
    //getIconInfo = GetDllProc("user32.dll", "GetIconInfo")

	errOutOfBounds = fmt.Errorf("index out of bounds")
)

type HICON Handle

type grpIconHeader struct { // size = 6 bytes
	Reserved uint16
	Type     uint16
	Count    uint16
}

type grpIconResEntry struct {
	Width        byte
	Height       byte
	ColorCount   byte
	Reserved     byte
	Planes       uint16
	BitCount     uint16
	BytesInRes   uint32
	ID           uint16
}

type iconFileEntry struct { // size = 16 bytes
	Width        byte
	Height       byte
	ColorCount   byte
	Reserved     byte
	Planes       uint16
	BitCount     uint16
	BytesInRes   uint32
	Offset       uint32
}

type iconHeader struct {
	header grpIconHeader
	entries []grpIconResEntry
	loadErr error
}

func (v *iconHeader) IsValid() bool {
	return v.loadErr == nil && v.header.Count > 0 && len(v.entries) == int(v.header.Count)
}

type Icon struct {
	module *ResourceModule
	resourceID *ResourceID
	resourceInfo *ResourceInfo
	header *iconHeader
	handle HICON
}

func newIcon(module *ResourceModule, id *ResourceID) (*Icon, error) {
	info, err := module.GetResourceInfo(id.GetValue(), windows.RT_GROUP_ICON)
	if err != nil {
		return nil, err
	}
	fmt.Println("New Icon created")
	return &Icon{module: module, resourceID : id, resourceInfo: info, handle: 0}, nil
}

func (v *Icon) GetHandle() (HICON, error) {
	if v.handle != 0 {
		return v.handle, nil
	}
    res, _, err := loadIcon.Call(uintptr(v.module.handle), v.resourceID.GetPtr())
    if res == 0 {
        return 0, fmt.Errorf("load Icon failed: %v", err)
    }
	v.handle = HICON(res)
    return v.handle, nil
}

func (v *Icon) headerNeeded() {
	if v.header != nil {
		return
	}
	v.header = new(iconHeader)
	b, err := v.resourceInfo.LoadData()
    if err != nil {
		v.header.loadErr = err
		return
    }
    reader := bytes.NewReader(b)
	if err := binary.Read(reader, binary.LittleEndian, &v.header.header); err != nil {
		v.header.loadErr = err
        return
	}
    v.header.entries = make([]grpIconResEntry, v.header.header.Count)
	for i := 0; i < int(v.header.header.Count); i++ {
		if err := binary.Read(reader, binary.LittleEndian, &v.header.entries[i]); err != nil {
			v.header.entries = v.header.entries[0:i]
			v.header.loadErr = err
			return
		}
	}
}

func (v *Icon) GetCount() int {
	v.headerNeeded()
	if v.header.IsValid() {
		return int(v.header.header.Count)
	}
	return 0
}

func (v *Icon) GetIconSize(index int) (width int, height int) {
	v.headerNeeded()
	if index < 0 || !v.header.IsValid() || int(v.header.header.Count) <= index {
		return 0, 0
	}
	return int(v.header.entries[index].Width), int(v.header.entries[index].Height) 
}

func (v *Icon) writeBytes(w io.Writer, entries []grpIconResEntry) error {
	v.headerNeeded()
	if !v.header.IsValid() {
		return v.header.loadErr
	}
	var h grpIconHeader
	h.Type = 1
	h.Count = uint16(len(entries))
	if err := binary.Write(w, binary.LittleEndian, &h); err != nil {
		return err
	}
	size := uint32(6 + len(entries) * 16)
	var re iconFileEntry
	for _, e := range(entries) {
		re.BitCount = e.BitCount
		re.BytesInRes = e.BytesInRes
		re.ColorCount = e.ColorCount
		re.Planes = e.Planes
		re.Width = e.Width
		re.Height = e.Height
		re.Offset = size
		if err := binary.Write(w, binary.LittleEndian, &re); err != nil {
			return err
		}
		size += e.BytesInRes
	}
	for _, e := range(entries) {
		if r, err := v.module.GetResourceInfo(e.ID, windows.RT_ICON); err != nil {
			return err
		} else {
			b, err := r.LoadData()
			if err != nil {
				return err
			}
			if _, err := w.Write(b); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *Icon) doWriteToFile(fileName string, entries[]grpIconResEntry) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	err = v.writeBytes(file, entries)
	if err != nil {
		file.Close()
		os.Remove(fileName)
		return err
	}
	return nil
}

func (v *Icon) getBytes(entries []grpIconResEntry) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := v.writeBytes(buf, entries); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *Icon) WriteToFile(fileName string, idx int) error {
	v.headerNeeded()
	if !v.header.IsValid() {
		return v.header.loadErr
	}
	if idx < 0 {
		return v.doWriteToFile(fileName, v.header.entries)
	}
	if idx >= int(v.header.header.Count) {
		return errOutOfBounds
	}
	return v.doWriteToFile(fileName, v.header.entries[idx:idx+1])
}

func (v *Icon) GetIconFileBytes(idx int) ([]byte, error) {
	v.headerNeeded()
	if !v.header.IsValid() {
		return nil, v.header.loadErr
	}
	if idx < 0 {
		return v.getBytes(v.header.entries)
	}
	if idx >= int(v.header.header.Count) {
		return nil, errOutOfBounds
	}
	return v.getBytes(v.header.entries[idx:idx+1])
}

func (v *Icon) searchNearest(cx, cy int) int {
	idx := -1
    var deltaX, deltaY int
    for i, e := range v.header.entries {
		dx := cx - int(e.Width)
		dy := cy - int(e.Height)
        if idx < 0 {
            idx = i
            deltaX, deltaY = dx, dy
        } else {
			if (dx > deltaX && (dx < deltaX || deltaX < 0)) || (dy > deltaY && (dy < deltaY || deltaY < 0)) {
                idx, deltaX, deltaY = i, dx, dy
            }
        }
    }
    return idx
}

func (v *Icon) SearchIcon(cx, cy int, nearest bool) int {
	v.headerNeeded()
	if !v.header.IsValid() {
		return -1
	}
	if !nearest {
		for i, e := range(v.header.entries) {
			if int(e.Width) == cx && int(e.Height) == cy {
				return i
			}
		}
	}
	return v.searchNearest(cx, cy)
}
