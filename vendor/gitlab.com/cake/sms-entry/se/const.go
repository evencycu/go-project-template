package se

import (
	"fmt"
	"runtime"
)

const (
	PathService = "service"

	RedisRateLimitDB   = 3
	RedisLCProvisionDB = 10
)

var (
	SubPathService = fmt.Sprintf("/service/:%s", PathService)
)

var (
	// SMS related APIs
	APIPathSMS         = "/v1/sms"
	APIPathInternalSMS = "/internal/v1/sms"
	// Phone related APIs
	APIPathPhone = "/v1/phone"
	// TDR related APIs
	APIPathTDR         = "/v1/tdr"
	APIPathInternalTDR = "/internal/v1/tdr"
	// Provision related
	APIPathProvision        = "/internal/v1"
	APIPathProvisionService = fmt.Sprintf("%s%s", APIPathProvision, SubPathService)
)

var (
	// appName is the service name for log
	appName = "sms-entry"
	// appVersion used for log and info
	appVersion = "unknown"
	gover      = runtime.Version()
	goos       = runtime.GOOS
	goarch     = runtime.GOARCH
	gitBranch  = "master"
	gitCommit  = "$Format:%H$"          // sha1 from git, output of $(git rev-parse HEAD)
	buildDate  = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

type Version struct {
	// Version is a binary version from git tag.
	Version string `json:"version"`
	// GitCommit is a git commit
	GitCommit string `json:"gitCommit"`
	// GitBranch is a git branch
	GitBranch string `json:"gitBranch"`
	// BuildDate is a build date of the binary.
	BuildDate string `json:"buildDate"`
	// GoOs holds OS name.
	GoOs string `json:"goOs"`
	// GoArch holds architecture name.
	GoArch string `json:"goArch"`
	// GoVer holds Golang build version.
	GoVer string `json:"goVer"`
}

// GetVersion returns version details.
func GetVersion() Version {
	return Version{
		appVersion,
		gitCommit,
		gitBranch,
		buildDate,
		goos,
		goarch,
		gover,
	}
}

func GetAppName() string {
	return appName
}
