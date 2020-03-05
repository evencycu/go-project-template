package se

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

var ip, pod, ns, env string

func GetNamespace() string {
	if ns == "" {
		ns = viper.GetString("app.namespace")
	}
	return ns
}

func GetPhaseEnv() string {
	if env == "" {
		ns := GetNamespace()
		env = "local"
		strs := strings.Split(ns, "-")
		if len(strs) > 0 {
			env = strs[0]
		}
	}
	return env
}

func GetPodName() string {
	if pod == "" {
		pod, _ = os.Hostname()
	}
	return pod
}

func GetIP() string {
	if ip == "" {
		ip = viper.GetString("app.pod_ip")
	}
	return ip
}
