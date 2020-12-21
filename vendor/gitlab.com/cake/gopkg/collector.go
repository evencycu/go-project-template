package gopkg

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/mod/modfile"
)

var (
	modFileContent = ""
)

// NewModFileCollector returns a collector which exports metrics about current dependency information from go.mod file.
func NewModFileCollector(program, goModPath string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_mod_file_info",
			Help: fmt.Sprintf(
				"A metric with a constant '1' value labeled by dependency name, version, from which %s's go.mod file.",
				program,
			),
		},
		[]string{"name", "version", "indirect", "replacedBy", "program"},
	)

	data, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return gauge
	}

	f, err := modfile.Parse("go.mod", data, func(_, version string) (string, error) { return version, nil })
	if err != nil {
		return gauge
	}

	replacedList := map[string]string{}
	for _, replaced := range f.Replace {
		replacedList[replaced.Old.String()] = replaced.New.String()
	}

	for _, pkg := range f.Require {
		replacedBy := "none"

		if v, ok := replacedList[pkg.Mod.String()]; ok {
			replacedBy = v
		}
		if v, ok := replacedList[pkg.Mod.Path]; ok {
			replacedBy = v
		}
		gauge.WithLabelValues(pkg.Mod.Path, pkg.Mod.Version, strconv.FormatBool(pkg.Indirect), replacedBy, program).Set(1)
	}

	return gauge
}
