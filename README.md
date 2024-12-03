# etcd module for caddy

This module implements request IP filter

## Features


## Caddyfile

```
:80 {
    reverse_proxy {
        etcd {
            endpoints http://localhost:2379
            key /services/my-service/backends
        }
    }

    log {
        output file /var/log/caddy/test.log
    }
}
```

## Parameters

| Name          | Description                                         | Type     | Default    |
| ------------- | --------------------------------------------------- | -------- | ---------- |
| interval      | Update Interval                                     | duration | 1h         |
| timeout       | Maximum time to wait to get a response from network | duration | no timeout |
| allow_ip_list | List of allowed IP addresses (Local path or URL)    | string   | none       |
| block_ip_list | List of blocked IP addresses (Local path or URL)    | string   | none       |
