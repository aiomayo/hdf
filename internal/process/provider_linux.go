package process

import (
	"syscall"

	gopsProcess "github.com/shirou/gopsutil/v4/process"
)

type linuxProvider struct{}

func New() Provider {
	return &linuxProvider{}
}

func (p *linuxProvider) List() ([]Info, error) {
	procs, err := gopsProcess.Processes()
	if err != nil {
		return nil, err
	}
	portMap := buildPortMap()
	result := make([]Info, 0, len(procs))
	for _, proc := range procs {
		info := procToInfo(proc, portMap)
		result = append(result, info)
	}
	return result, nil
}

func (p *linuxProvider) FindByPID(pid int32) (*Info, error) {
	proc, err := gopsProcess.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	portMap := buildPortMap()
	info := procToInfo(proc, portMap)
	return &info, nil
}

func (p *linuxProvider) FindByPort(port uint32) ([]Info, error) {
	portMap := buildPortMap()
	procs, err := gopsProcess.Processes()
	if err != nil {
		return nil, err
	}
	var result []Info
	for _, proc := range procs {
		pid := proc.Pid
		if portMap[pid] == port {
			info := procToInfo(proc, portMap)
			result = append(result, info)
		}
	}
	return result, nil
}

func (p *linuxProvider) Children(pid int32) ([]Info, error) {
	proc, err := gopsProcess.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	children, err := proc.Children()
	if err != nil {
		return nil, err
	}
	portMap := buildPortMap()
	result := make([]Info, 0, len(children))
	for _, child := range children {
		info := procToInfo(child, portMap)
		result = append(result, info)
	}
	return result, nil
}

func (p *linuxProvider) Kill(pid int32) error {
	return syscall.Kill(int(pid), syscall.SIGKILL)
}

func (p *linuxProvider) Terminate(pid int32) error {
	return syscall.Kill(int(pid), syscall.SIGTERM)
}

func (p *linuxProvider) Signal(pid int32, sig Signal) error {
	return syscall.Kill(int(pid), syscall.Signal(sig))
}

func (p *linuxProvider) IsRunning(pid int32) bool {
	proc, err := gopsProcess.NewProcess(pid)
	if err != nil {
		return false
	}
	running, err := proc.IsRunning()
	if err != nil {
		return false
	}
	return running
}
