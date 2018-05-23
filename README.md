[![Build Status](https://travis-ci.com/kontena/pharos-host-upgrades.svg?token=pcXcAqCByqv2epJ6v1zo&branch=master)](https://travis-ci.com/kontena/pharos-host-upgrades)

# pharos-host-upgrades

See [`resources`](./resources) for example kube manifests.

## Usage

```
Usage of pharos-host-upgrades:
  -alsologtostderr
    	log to standard error as well as files
  -kube-daemonset string
    	Name of kube DaemonSet (KUBE_DAEMONSET)
  -kube-namespace string
    	Name of kube Namespace (KUBE_NAMESPACE)
  -kube-node string
    	Name of kube Node (KUBE_NODE)
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -schedule string
    	Scheduled upgrade (cron syntax)
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
```

### `--schedule`

See https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format

Variant of a standard crontab with a leading seconds field.

Examples:

* `0 15 5 * * *` - every day at 05:15:00

## Development

Using the vagrant machines:

    $ vagrant up ubuntu
    $ vagrant up centos-7

### Setup

    $ mkdir .kube
    $ sudo cat /etc/kubernetes/admin.conf > .kube/config

### Build

    /vagrant $ docker build -t kontena/pharos-host-upgrades:dev .

### Test

    /vagrant $ docker run --rm --name host-upgrades --privileged \
      -v $PWD/config:/etc/host-upgrades \
      -v /run/host-upgrades:/run/host-upgrades \
      -v /var/run/dbus:/var/run/dbus \
      -v /run/log/journal:/run/log/journal \
      kontena/pharos-host-upgrades:dev
