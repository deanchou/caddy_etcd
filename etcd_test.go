package caddy_etcd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestDefault(t *testing.T) {
	endpoints := []string{"192.168.31.200:2379", "127.0.0.1:2379"}
	key := "/buoy/buoy.module_pak"
	config := fmt.Sprintf(`etcd {
            endpoints %s
            key %s
        }`, strings.Join(endpoints, " "), key)

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

	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	err = r.Provision(ctx)
	if err != nil {
		t.Errorf("error provisioning %q: %v", config, err)
	}

	if r.client == nil {
		t.Errorf("expected client to be non-nil")
	}

	backends, err := r.GetBackends()
	if err != nil {
		t.Errorf("error getting backends: %v", err)
	}

	if len(backends) == 0 {
		t.Errorf("expected some backends, got none")
	}
}
