package apps

import (
	"fmt"
	"strings"
	"time"

	"golift.io/starr"
	"golift.io/starr/prowlarr"
)

// prowlarrHandlers is called once on startup to register the web API paths.
func (a *Apps) prowlarrHandlers() {
}

// ProwlarrConfig represents the input data for a Prowlarr server.
type ProwlarrConfig struct {
	starrConfig
	*starr.Config
	*prowlarr.Prowlarr `toml:"-" xml:"-" json:"-"`
	errorf             func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupProwlarr(timeout time.Duration) error {
	for i, prowl := range a.Prowlarr {
		if prowl.Config == nil || prowl.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Prowlarr config %d", ErrInvalidApp, i+1)
		}

		prowl.Debugf = a.DebugLog.Printf
		prowl.errorf = a.ErrorLog.Printf
		prowl.setup(timeout)
	}

	return nil
}

func (r *ProwlarrConfig) setup(timeout time.Duration) {
	r.Prowlarr = prowlarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}

	r.URL = strings.TrimRight(r.URL, "/")

	if u, err := r.GetURL(); err != nil {
		r.errorf("Checking Prowlarr Path: %v", err)
	} else if u := strings.TrimRight(u, "/"); u != r.URL {
		r.errorf("Prowlarr URL fixed: %s -> %s (continuing)", r.URL, u)
		r.URL = u
	}
}
