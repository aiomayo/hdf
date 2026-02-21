package process

import (
	"time"

	gopsNet "github.com/shirou/gopsutil/v4/net"
	gopsProcess "github.com/shirou/gopsutil/v4/process"
)

func procToInfo(proc *gopsProcess.Process, portMap map[int32]uint32) Info {
	pid := proc.Pid
	name, _ := proc.Name()
	cmdline, _ := proc.Cmdline()
	user, _ := proc.Username()
	ppid, _ := proc.Ppid()
	cpu, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()
	createMs, _ := proc.CreateTime()

	var memRSS uint64
	if memInfo != nil {
		memRSS = memInfo.RSS
	}

	var children []int32
	if ch, err := proc.Children(); err == nil {
		for _, c := range ch {
			children = append(children, c.Pid)
		}
	}

	return Info{
		PID:        pid,
		PPID:       ppid,
		Name:       name,
		Cmdline:    cmdline,
		User:       user,
		Port:       portMap[pid],
		CPUPercent: cpu,
		MemRSS:     memRSS,
		CreateTime: time.UnixMilli(createMs),
		Children:   children,
	}
}

func buildPortMap() map[int32]uint32 {
	portMap := make(map[int32]uint32)
	conns, err := gopsNet.Connections("all")
	if err != nil {
		return portMap
	}
	for _, conn := range conns {
		if conn.Status == "LISTEN" && conn.Pid != 0 {
			if _, exists := portMap[conn.Pid]; !exists {
				portMap[conn.Pid] = conn.Laddr.Port
			}
		}
	}
	return portMap
}
