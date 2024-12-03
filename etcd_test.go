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
	endpoints := []string{"http://192.168.31.200:2379"}
	key := "/buoy/buoy.core"
	version := "1.0.0"
	config := fmt.Sprintf(`etcd {
            endpoints %s
            key %s
			version %s
			timeout 5s
        }`, strings.Join(endpoints, " "), key, version)

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

	if r.Version != version {
		t.Errorf("expected version to be %q, got %q", version, r.Version)
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
