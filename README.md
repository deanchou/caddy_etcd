# etcd module for caddy

This module implements reverse_proxy upstreams from etcd

## Features

- dynamic upstreams
- version match

## Caddyfile

```
:80 {
    reverse_proxy {
       dynamic etcd {
            endpoints http://localhost:2379
            key /services/my-service/backends
			version_key /services/my-service/backends.version
			timeout 5s
        }
        lb_policy least_conn
    }

    log {
        output file /var/log/caddy/test.log
    }
}
```

## Parameters

| Name        | Description                                     | Type         | Default     |
| ----------- | ----------------------------------------------- | ------------ | ----------- |
| endpoints   | etcd endpoints                                  | string array | none        |
| key         | get upstreams with key from etcd                | string       | none        |
| version_key | version match with key from etcd                | string       | key.version |
| timeout     | Maximum time to wait to get upstreams from etcd | duration     | no timeout  |
