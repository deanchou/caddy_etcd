package caddy_etcd

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/tidwall/gjson"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(EtcdProxy{})
}

// EtcdProxy is a Caddy module that integrates etcd with reverse_proxy.
type EtcdProxy struct {
	Endpoints  []string       `json:"endpoints,omitempty"`
	Key        string         `json:"key,omitempty"`
	VersionKey string         `json:"version_key,omitempty"`
	Timeout    caddy.Duration `json:"timeout,omitempty"`

	client *clientV3.Client
	ctx    caddy.Context
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (EtcdProxy) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.reverse_proxy.upstreams.etcd",
		New: func() caddy.Module { return new(EtcdProxy) },
	}
}

// Provision sets up the module.
func (ep *EtcdProxy) Provision(ctx caddy.Context) error {
	ep.ctx = ctx
	ep.logger = ctx.Logger(ep)

	if ep.VersionKey == "" {
		ep.VersionKey = ep.Key + ".version"
	}

	cli, err := clientV3.New(clientV3.Config{
		Endpoints:   ep.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		ep.logger.Error("failed to create etcd client", zap.Error(err))
		return err
	}

	ep.client = cli
	return nil
}

func (ep *EtcdProxy) GetUpstreams(r *http.Request) ([]*reverseproxy.Upstream, error) {
	ctx, cancel := ep.getContext()
	defer cancel()

	var version string

	kv := clientV3.NewKV(ep.client)

	if ep.VersionKey != "" {
		resp, err := kv.Get(ctx, ep.VersionKey)
		if err != nil {
			ep.logger.Error("failed to get key from etcd", zap.Error(err))
			return nil, err
		}

		if len(resp.Kvs) > 0 {
			version = string(resp.Kvs[0].Value)
		}
	}

	resp, err := kv.Get(ctx, ep.Key, clientV3.WithPrefix())
	if err != nil {
		ep.logger.Error("failed to get key from etcd", zap.Error(err))
		return nil, err
	}

	upstreams := make([]*reverseproxy.Upstream, 0, 1)
	for _, kv := range resp.Kvs {
		ep.logger.Info("got key from etcd", zap.String("key", string(kv.Key)), zap.String("value", string(kv.Value)))

		if version != "" {
			v := gjson.GetBytes(kv.Value, "version").String()
			if version != v {
				continue
			}
		}

		endpoints := gjson.GetBytes(kv.Value, "endpoints").Array()
		for _, endpoint := range endpoints {
			dial := endpoint.String()
			dial = dial[strings.Index(dial, "://")+3:]
			upstreams = append(upstreams, &reverseproxy.Upstream{
				Dial: dial,
			})
		}
	}
	ep.logger.Info("got upstreams from etcd", zap.Any("upstreams", upstreams))

	return upstreams, nil
}

// UnmarshalCaddyfile sets up the module from Caddyfile tokens.
func (ep *EtcdProxy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "endpoints":
				ep.Endpoints = d.RemainingArgs()
				if len(ep.Endpoints) == 0 {
					return d.ArgErr()
				}
			case "key":
				if !d.Args(&ep.Key) {
					return d.ArgErr()
				}
			case "version_key":
				if !d.Args(&ep.VersionKey) {
					return d.ArgErr()
				}
			case "timeout":
				if !d.NextArg() {
					return d.ArgErr()
				}
				val, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				ep.Timeout = caddy.Duration(val)
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	return nil
}

// getContext
func (ep *EtcdProxy) getContext() (context.Context, context.CancelFunc) {
	if ep.Timeout > 0 {
		return context.WithTimeout(ep.ctx, time.Duration(ep.Timeout))
	}
	return context.WithCancel(ep.ctx)
}

var (
	_ caddy.Module                = (*EtcdProxy)(nil)
	_ caddy.Provisioner           = (*EtcdProxy)(nil)
	_ caddyfile.Unmarshaler       = (*EtcdProxy)(nil)
	_ reverseproxy.UpstreamSource = (*EtcdProxy)(nil)
)
