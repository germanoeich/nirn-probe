# Nirn-probe
Nirn-probe is a small probe that queries Discord periodically (every 1 minute) and checks if a machine/kubernetes node is banned by Cloudflare. It notifies you using Prometheus metrics.

Having issues with Discord ratelimits? [Nirn-proxy](https://github.com/germanoeich/nirn-proxy) might be useful for you.
#### Features
- Support for VMs or Kubernetes nodes
- Extremely lightweight
- No config
- Extremely cheap on the Prometheus side

### Usage
Binaries can be found [here](https://github.com/germanoeich/nirn-probe/releases). Docker images can be found [here](https://github.com/germanoeich/nirn-probe/pkgs/container/nirn-probe)

Configuration options are

| Variable        | Value  | Default |
|-----------------|--------|---------|
| LOG_LEVEL       | panic, fatal, error, warn, info, debug, trace | info |
| PORT            | number | 8100    |
| NODE_NAME       | string | os.Hostname() |

.env files are loaded if present

### Behaviour

Once started, the probe starts firing requests to /api/v9/gateway. This endpoint can be accesed without a token and has no ratelimits. If the probe encounters a 429, the `nirn_probe_banned` gauge will be set to 1. It will remain as 1 until the probe encounters another status code that is not equal to 429.

The probe also exports the `nirn_probe_up` metric, which is always set to 1 and allows you to alert if the probe is offline.

NODE_NAME is attached to every metric as a label called "node". It defaults to the hostname in the system, and in the included Kubernetes config, it defaults to the kubernetes node name.

### Kubernetes

The probe was designed to run inside of Kubernetes as a DaemonSet. You can find a pre-baked YAML [here](https://github.com/germanoeich/nirn-probe/blob/main/kubernetes/nirn-probe-daemonset.yaml), which includes the DaemonSet, a Service and a prometheus-operator ServiceMonitor. It also sets the NODE_NAME env to the k8s node name.

### Metrics

| Key               | Labels | Description                                    |
|-------------------|--------|------------------------------------------------|
|nirn_probe_error   | node   | Counter for errors                             |
|nirn_probe_banned  | node   | Gauge - 0 = ok, 1 = banned                     |
|nirn_probe_up      | node   | Gauge - always 1                               |

### Dashboard

A grafana dashboard can be found [here](https://github.com/germanoeich/nirn-probe/blob/main/grafana/dash.json). It looks like [this](https://prnt.sc/23x5q9l)