package haproxy

type HAProxyHelperConfig struct {
	HAProxyConfigTemplate string `json:"haproxyConfigTemplate"`
	HAProxyConfigPath     string `json:"haproxyConfigPath"`
	ReloadCommand         string `json:"reloadCommand"`
	HTTPInternalPort      int    `json:"httpInternalPort"`
	HTTPSInternalPort     int    `json:"httpsInternalPort"`
}

const haproxyConfigTemplate = `
global
    maxconn     50000
    log         127.0.0.1 local0
    user        haproxy
    daemon

frontend http
    bind :80
    bind :::80
    mode tcp
    default_backend     http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend https
    bind :443
    bind :::443
    mode tcp
    default_backend     https
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

backend http
    mode        tcp
    balance     roundrobin
{{- range .Nodes }}
    server {{ .Name }} {{ .InternalIP }}:{{ $.HTTPInternalPort }} check
{{- end }}

backend https
    mode        tcp
    balance     roundrobin
{{- range .Nodes }}
    server {{ .Name }} {{ .InternalIP }}:{{ $.HTTPSInternalPort }} check
{{- end }}
`

var (
	defaultHAProxyHelperConfig = &HAProxyHelperConfig{
		HAProxyConfigTemplate: haproxyConfigTemplate,
		HAProxyConfigPath:     "/etc/haproxy/haproxy.cfg",
		ReloadCommand:         "systemctl reload haproxy",
		HTTPInternalPort:      80,
		HTTPSInternalPort:     443,
	}
)
