package m800log

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/cake/goctx"
)

const (
	// BuildInFieldNumber defines the number of m800 log built-in log fields
	BuildInFieldNumber = 8
)

// M800JSONFormatter formats logs into the m800 log style json.
type M800JSONFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
	App             string
	Version         string
	Host            string
	Env             string
	Namespace       string
}

func newM800JSONFormatter(timestampFormat, app, version, env, ns string) *M800JSONFormatter {
	if timestampFormat == "" {
		timestampFormat = time.RFC3339Nano
	}
	host, _ := os.Hostname()
	return &M800JSONFormatter{
		TimestampFormat: timestampFormat,
		App:             app,
		Version:         version,
		Host:            host,
		Env:             env,
		Namespace:       ns,
	}
}

// Format renders a single log entry
func (f *M800JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	delete(entry.Data, goctx.LogKeyTrace)
	data := make(logrus.Fields, len(entry.Data)+BuildInFieldNumber)
	for k, v := range entry.Data {
		data[k] = v
	}

	data[goctx.LogKeyApp] = f.App
	data[goctx.LogKeyTimestamp] = entry.Time.UTC().Format(f.TimestampFormat)
	data[goctx.LogKeyMessage] = entry.Message
	data[goctx.LogKeyLevel] = entry.Level.String()
	data[goctx.LogKeyVersion] = f.Version
	data[goctx.LogKeyInstance] = f.Host
	data[goctx.LogKeyNamespace] = f.Namespace
	data[goctx.LogKeyEnv] = f.Env

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
