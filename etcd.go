package caddy_etcd

import (
	"context"
	"net/http"

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
	Endpoints []string `json:"endpoints,omitempty"`
	Key       string   `json:"key,omitempty"`
	client    *clientV3.Client

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

	cli, err := clientV3.New(clientV3.Config{
		Endpoints: ep.Endpoints,
	})
	if err != nil {
		ep.logger.Error("failed to create etcd client", zap.Error(err))
		return err
	}
	ep.client = cli
	return nil
}

func (ep *EtcdProxy) GetUpstreams(r *http.Request) ([]*reverseproxy.Upstream, error) {
	upstreams := make([]*reverseproxy.Upstream, 0, 1)

	resp, err := ep.client.Get(context.Background(), ep.Key, clientV3.WithPrefix())
	if err != nil {
		ep.logger.Error("failed to get key from etcd", zap.Error(err))
		return nil, err
	}
	ep.logger.Info("got backends from etcd", zap.Any("resp", resp))

	for _, kv := range resp.Kvs {
		endpoints := gjson.GetBytes(kv.Value, "endpoints").Array()
		for _, endpoint := range endpoints {
			upstreams = append(upstreams, &reverseproxy.Upstream{
				Dial: endpoint.String(),
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
			case "key":
				if !d.Args(&ep.Key) {
					return d.ArgErr()
				}
			}
		}
	}
	return nil
}

var (
	_ caddy.Module                = (*EtcdProxy)(nil)
	_ caddy.Provisioner           = (*EtcdProxy)(nil)
	_ caddyfile.Unmarshaler       = (*EtcdProxy)(nil)
	_ reverseproxy.UpstreamSource = (*EtcdProxy)(nil)
)
