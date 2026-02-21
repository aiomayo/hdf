package config

import "time"

var defaultProtected = []string{
	"init",
	"systemd",
	"launchd",
	"kernel_task",
	"WindowServer",
	"loginwindow",
	"sshd",
}

var defaultConfig = Config{
	GracefulTimeout: 5 * time.Second,
	Protected:       defaultProtected,
	Aliases:         map[string]string{},
	DefaultForce:    false,
	DefaultVerbose:  false,
}
