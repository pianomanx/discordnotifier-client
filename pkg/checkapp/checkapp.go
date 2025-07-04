// Package checkapp provides a suite of small procedures to check integration URLs and commands.
// This is used by all the double-green check marks on the client UI.
package checkapp

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/gorilla/mux"
	"golift.io/cnfgfile"
)

const (
	success    = "Connection Successful! Version: "
	connecting = "Connecting: "
	validation = "Validation: "
)

type Input struct {
	Real  *configfile.Config
	Post  *configfile.Config
	Type  string
	Args  url.Values
	Index int
}

var ErrBadIndex = errors.New("index provided has no configuration data")

func Test(orig *configfile.Config, writer http.ResponseWriter, req *http.Request) {
	posted := configfile.Config{}

	if err := json.NewDecoder(req.Body).Decode(&posted); err != nil {
		http.Error(writer, "Decoding JSON data into Go data structure failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := cnfgfile.Parse(&posted, &cnfgfile.Opts{
		Name:          mnd.Title,
		TransformPath: configfile.ExpandHomedir,
		Prefix:        "filepath:",
	}); err != nil {
		http.Error(writer, "Parsing extra config filepaths: "+err.Error(), http.StatusBadRequest)
		return
	}

	index, _ := strconv.Atoi(mux.Vars(req)["index"])
	reply, code := testInstance(req.Context(), &Input{
		Real:  orig,
		Post:  &posted,
		Type:  mux.Vars(req)["type"],
		Args:  req.URL.Query(),
		Index: index,
	})
	http.Error(writer, html.EscapeString(reply), code)
}

//nolint:funlen,cyclop // It's really not that bad.
func testInstance(ctx context.Context, input *Input) (string, int) {
	switch strings.ToLower(input.Type) {
	// commands.go
	case "commands":
		return testCommand(ctx, input)
		// endpoints.go
	case "endpoints":
		return testEndpoint(ctx, input)
	// downloaders.go
	case "nzbget":
		return checkAndRun(ctx, testNZBGet, input, input.Post.AppsConfig, input.Post.NZBGet)
	case "deluge":
		return checkAndRun(ctx, testDeluge, input, input.Post.AppsConfig, input.Post.Deluge)
	case "qbit", "qbittorrent":
		return checkAndRun(ctx, testQbit, input, input.Post.AppsConfig, input.Post.Qbit)
	case "rtorrent":
		return checkAndRun(ctx, testRtorrent, input, input.Post.AppsConfig, input.Post.Rtorrent)
	case "transmission":
		return checkAndRun(ctx, testTransmission, input, input.Post.AppsConfig, input.Post.Transmission)
	case "sabnzb", "sabnzbd":
		return checkAndRun(ctx, testSabNZB, input, input.Post.AppsConfig, input.Post.SabNZB)
	// starr.go
	case "lidarr":
		return checkAndRun(ctx, testLidarr, input, input.Post.AppsConfig, input.Post.Lidarr)
	case "prowlarr":
		return checkAndRun(ctx, testProwlarr, input, input.Post.AppsConfig, input.Post.Prowlarr)
	case "radarr":
		return checkAndRun(ctx, testRadarr, input, input.Post.AppsConfig, input.Post.Radarr)
	case "readarr":
		return checkAndRun(ctx, testReadarr, input, input.Post.AppsConfig, input.Post.Readarr)
	case "sonarr":
		return checkAndRun(ctx, testSonarr, input, input.Post.AppsConfig, input.Post.Sonarr)
	// snapshots.go
	case "mysql":
		return checkAndRun(ctx, testMySQL, input, input.Post.Snapshot, input.Post.Snapshot.MySQL)
	case "nvidia":
		return checkAndRun(ctx, testNvidia, input, input.Post.Snapshot,
			[]snapshot.NvidiaConfig{input.Post.Snapshot.Nvidia}) // ad-hoc slice, index is already 0.
	// services.go
	case "tcp":
		return checkAndRun(ctx, testTCP, input, input.Post.Service, input.Post.Service)
	case "http":
		return checkAndRun(ctx, testHTTP, input, input.Post.Service, input.Post.Service)
	case "process":
		return checkAndRun(ctx, testProcess, input, input.Post.Service, input.Post.Service)
	case "ping", "icmp":
		return checkAndRun(ctx, testPing, input, input.Post.Service, input.Post.Service)
	// media.go
	case "plex":
		return testPlex(ctx, input.Post.Plex)
	case "tautulli":
		return testTautulli(ctx, input.Post.Tautulli)
	default:
		return "Unknown Check Type Requested! (" + input.Type + ")", http.StatusNotImplemented
	}
}

// checkAndRun makes sure the slice length is at least as long as the index, and checks parents for nil.
func checkAndRun[D any](
	ctx context.Context,
	checker func(ctx context.Context, input D) (string, int),
	input *Input,
	parent any,
	slice []D,
) (string, int) {
	if parent == nil || len(slice) <= input.Index {
		return ErrBadIndex.Error(), http.StatusBadRequest
	}

	return checker(ctx, slice[input.Index])
}
