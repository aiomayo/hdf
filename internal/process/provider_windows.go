package process

import (
	gopsProcess "github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
)

type windowsProvider struct{}

func New() Provider {
	return &windowsProvider{}
}

func (p *windowsProvider) List() ([]Info, error) {
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

func (p *windowsProvider) FindByPID(pid int32) (*Info, error) {
	proc, err := gopsProcess.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	portMap := buildPortMap()
	info := procToInfo(proc, portMap)
	return &info, nil
}

func (p *windowsProvider) FindByPort(port uint32) ([]Info, error) {
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

func (p *windowsProvider) Children(pid int32) ([]Info, error) {
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

func (p *windowsProvider) Kill(pid int32) error {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	return windows.TerminateProcess(handle, 1)
}

func (p *windowsProvider) Terminate(pid int32) error {
	return p.Kill(pid)
}

func (p *windowsProvider) Signal(pid int32, _ Signal) error {
	return p.Kill(pid)
}

func (p *windowsProvider) IsRunning(pid int32) bool {
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
