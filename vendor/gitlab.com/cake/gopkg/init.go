package gopkg

import (
	"runtime/debug"

	"github.com/povilasv/prommod"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	buildInfo, ok = debug.ReadBuildInfo()
)

func init() {
	prometheus.MustRegister(NewBuildInfoCollector())
	prometheus.MustRegister(prommod.NewCollector(appName))
	prometheus.MustRegister(NewModFileCollector(appName, "./go.mod"))
}

// NewBuildInfoCollector collects build info
func NewBuildInfoCollector() *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_build_info",
			Help: "Build information about the main Go module.",
		},
		[]string{"path", "version", "revision"},
	)
	if !ok {
		return gauge
	}

	gauge.WithLabelValues(buildInfo.Main.Path, appVersion, gitCommit).Set(1)

	return gauge
}
