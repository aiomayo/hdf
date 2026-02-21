package process

import "time"

type Info struct {
	PID        int32
	PPID       int32
	Name       string
	Cmdline    string
	User       string
	Port       uint32
	CPUPercent float64
	MemRSS     uint64
	CreateTime time.Time
	Children   []int32
}
