package process

import "syscall"

type Signal syscall.Signal

const (
	SignalTerm Signal = Signal(syscall.SIGTERM)
	SignalKill Signal = Signal(syscall.SIGKILL)
	SignalInt  Signal = Signal(syscall.SIGINT)
	SignalHup  Signal = Signal(syscall.SIGHUP)
)

type Provider interface {
	List() ([]Info, error)
	FindByPID(pid int32) (*Info, error)
	FindByPort(port uint32) ([]Info, error)
	Children(pid int32) ([]Info, error)
	Kill(pid int32) error
	Terminate(pid int32) error
	Signal(pid int32, sig Signal) error
	IsRunning(pid int32) bool
}
