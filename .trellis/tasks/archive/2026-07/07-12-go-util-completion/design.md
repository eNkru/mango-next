# Design: go-util-completion

## Proxy

```go
&http.Client{
  Timeout: ...,
  Transport: &http.Transport{
    Proxy: http.ProxyFromEnvironment,
  },
}
```

Apply in:
- `plugin.NewSandbox` (DefaultClient → configured client)
- `plugin.NewDownloader` httpClient

Go’s ProxyFromEnvironment already implements NO_PROXY hostname matching
(close enough to Crystal split-by-comma equality; document minor differences).

## Validation / web

Document only in gap-report; no new packages required.
