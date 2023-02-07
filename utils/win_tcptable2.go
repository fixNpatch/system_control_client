package utils

import (
	"diplom_client/structs"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/windows"
	"syscall"
	"time"
	"unsafe"
)

const (
	iphlpapiDll         = "iphlpapi.dll"
	tcpFnExtended       = "GetExtendedTcpTable"
	udpFn               = "GetExtendedUdpTable"
	setTCPEntry         = "SetTCPEntry"
	tcpTableOwnerPidAll = 5
)

func GetExtendedTCPTable() ([]structs.Connection, error) {
	result := make([]structs.Connection, 0)

	moduleHandle, err := windows.LoadLibrary(iphlpapiDll)
	if err != nil {
		return nil, err
	}

	ptr, err := windows.GetProcAddress(moduleHandle, tcpFnExtended)
	if err != nil {
		return nil, err
	}

	res, err := getNetTable(ptr, true, windows.AF_INET, tcpTableOwnerPidAll)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Cannot get TCPTable.", err))
	}

	if res != nil && len(res) >= 4 {
		count := *(*uint32)(unsafe.Pointer(&res[0]))
		const structLen = 24
		for n, pos := uint32(0), 4; n < count && pos+structLen <= len(res); n, pos = n+1, pos+structLen {
			state := *(*uint32)(unsafe.Pointer(&res[pos]))
			if state < 1 || state > 12 {
				return nil, errors.New(fmt.Sprint("State", state, "on", res[pos+4:pos+8], "unsupported"))
			}

			laddr := res[pos+4 : pos+8]
			raddr := res[pos+12 : pos+16]

			p := process.Process{Pid: int32(*(*uint32)(unsafe.Pointer(&res[pos+20])))}
			puser, _ := p.Username()
			pname, _ := p.Name()

			result = append(result, structs.Connection{
				LAddr: []int{
					int(laddr[0]),
					int(laddr[1]),
					int(laddr[2]),
					int(laddr[3]),
				},
				LPort: int(binary.BigEndian.Uint16(res[pos+8 : pos+10])),
				RAddr: []int{
					int(raddr[0]),
					int(raddr[1]),
					int(raddr[2]),
					int(raddr[3]),
				},
				RPort:       int(binary.BigEndian.Uint16(res[pos+16 : pos+18])),
				Pid:         int(*(*uint32)(unsafe.Pointer(&res[pos+20]))),
				ProcName:    pname,
				ProcOwner:   puser,
				ActiveSince: time.Now(),
				Status:      1,
			})
		}
	} else {
		return nil, nil
	}
	return result, nil
}

func getNetTable(fn uintptr, order bool, family int, class int) ([]byte, error) {
	var sorted uintptr
	if order {
		sorted = 1
	}
	for size, ptr, addr := uint32(8), []byte(nil), uintptr(0); ; {
		err, _, _ := syscall.Syscall6(fn, 5, addr, uintptr(unsafe.Pointer(&size)), sorted, uintptr(family), uintptr(class), 0)
		if err == 0 {
			return ptr, nil
		} else if err == uintptr(syscall.ERROR_INSUFFICIENT_BUFFER) {
			ptr = make([]byte, size)
			addr = uintptr(unsafe.Pointer(&ptr[0]))
		} else {
			return nil, fmt.Errorf("getNetTable failed: %v", err)
		}
	}
}
