package killer

type Action int

const (
	ActionTerminate Action = iota
	ActionKill
	ActionGraceful
)

func (a Action) String() string {
	switch a {
	case ActionTerminate:
		return "terminate"
	case ActionKill:
		return "kill"
	case ActionGraceful:
		return "graceful"
	default:
		return "unknown"
	}
}
