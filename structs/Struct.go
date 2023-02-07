package structs

import (
	"golang.org/x/net/websocket"
	"time"
)

// Mode
const (
	SaveOnlyChanges = iota
	SaveFullData
)

const (
	CloseEvent = iota
	RefreshConnectionListEvent
)

type Event struct {
	Name  int
	Data  interface{}
	Delay time.Time
}

type VmStat struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	Free        uint64
	Active      uint64
	Inactive    uint64
	Wired       uint64
}

type Procs struct {
	Pid        int32
	Name       string
	Status     string
	ParentPid  int32
	Uids       []int32
	Gids       []int32
	Groups     []int32
	NumThreads int32
	CreateTime int64
}

type Cpu struct {
	Percentage []float64
	Model      string
	Cores      int
}
type Disk struct {
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

type HostInfo struct {
	Procs           uint64
	OS              string
	PlatformVersion string
	Platform        string
}

type Host struct {
	Id     int
	Name   string
	IP     string
	Status int
}

type Connection struct {
	// Так как соединение может быть активно долгое время, то для него нет однозначного соответствия id - connection.
	// Поэтому FakeID
	FakeId      int    // id of connection.
	LAddr       []int  // local address
	LPort       int    // local port
	RAddr       []int  // remote address
	RPort       int    // remote port
	Pid         int    // process id
	ProcName    string // process name
	ProcOwner   string // process owner
	ActiveSince time.Time
	ClosedWhen  time.Time
	Status      int
}

type Stats struct {
	VmStat      VmStat
	Disk        Disk
	Cpu         Cpu
	Host        HostInfo
	Processes   []Procs
	Connections []Connection
	HostTime    time.Time
}

type Conn struct {
	conn   *websocket.Conn
	status int
}
