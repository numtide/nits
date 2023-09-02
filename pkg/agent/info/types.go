package info

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type Request struct {
	All    bool `json:"all"`
	Host   bool `json:"host"`
	Cpus   bool `json:"cpus"`
	Load   bool `json:"load"`
	Memory bool `json:"memory"`
	Disk   bool `json:"disk"`
}

type Response struct {
	NKey    string         `json:"nkey"`
	Name    string         `json:"name"`
	Subject string         `json:"subject"`
	Host    *host.InfoStat `json:"host,omitempty"`
	Cpus    []cpu.InfoStat `json:"cpus,omitempty"`
	Load    *Load          `json:"load,omitempty"`
	Memory  *Memory        `json:"memory,omitempty"`
	Disk    *Disk          `json:"disk,omitempty"`

	LastSeen time.Time
}

type Load struct {
	Avg  *load.AvgStat  `json:"avg,omitempty"`
	Misc *load.MiscStat `json:"misc,omitempty"`
}

type Memory struct {
	Virtual     *mem.VirtualMemoryStat `json:"virtual,omitempty"`
	Swap        *mem.SwapMemoryStat    `json:"swap,omitempty"`
	SwapDevices []*mem.SwapDevice      `json:"swapDevices,omitempty"`
}

type Disk struct {
	Partitions []disk.PartitionStat `json:"partitions,omitempty"`
}
