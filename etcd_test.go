package caddy_etcd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestDefault(t *testing.T) {
	endpoints := []string{"http://127.0.0.1:2379"}
	key := "/buoy/buoy.core"
	versionKey := "/buoy/buoy.core.version"
	config := fmt.Sprintf(`etcd {
            endpoints %s
            key %s
			version_key %s
			timeout 5s
        }`, strings.Join(endpoints, " "), key, versionKey)

	d := caddyfile.NewTestDispenser(config)

	r := EtcdProxy{}
	err := r.UnmarshalCaddyfile(d)
	if err != nil {
		t.Errorf("unmarshal error for %q: %v", config, err)
		return
	}

	if r.Endpoints == nil {
		t.Errorf("expected endpoints to be %q, got %q", endpoints, r.Endpoints)
	}

	if r.Key != key {
		t.Errorf("expected key to be %q, got %q", key, r.Key)
	}

	if r.VersionKey != versionKey {
		t.Errorf("expected version to be %q, got %q", versionKey, r.VersionKey)
	}

	if r.Timeout != caddy.Duration(5*time.Second) {
		t.Errorf("expected timeout to be 5s, got %v", r.Timeout)
	}

	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	err = r.Provision(ctx)
	if err != nil {
		t.Errorf("error provisioning %q: %v", config, err)
	}

	if r.client == nil {
		t.Errorf("expected client to be non-nil")
	}

	backends, err := r.GetUpstreams(nil)
	if err != nil {
		t.Errorf("error getting backends: %v", err)
	}

	if len(backends) == 0 {
		t.Errorf("expected some backends, got none")
	}
}
