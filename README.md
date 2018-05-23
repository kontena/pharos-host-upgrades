[![Build Status](https://travis-ci.com/kontena/pharos-host-upgrades.svg?token=pcXcAqCByqv2epJ6v1zo&branch=master)](https://travis-ci.com/kontena/pharos-host-upgrades)

# pharos-host-upgrades

See [`resources`](./resources) for example kube manifests.

### Supported host OS configurations

#### Ubuntu 16.04

    apt-get install unattended-upgrades

Disable `apt-periodic`:

```
sed -i 's/APT::Periodic::Unattended-Upgrade .*/APT::Periodic::Unattended-Upgrade "0";/' /etc/apt/apt.conf.d/20auto-upgrades
```

Alternatively:

    systemctl stop apt-daily.timer apt-daily-upgrade.timer
    systemctl disable apt-daily.timer apt-daily-upgrade.timer

#### CentOS 7

    yum install yum-cron

Ensure that the `yum-cron` service is stopped and disabled, as this will interfere with `pharos-host-upgrades`:

    systemctl stop yum-cron.service
    systemctl disable yum-cron.service

## Usage

#### Native

The Go binary can be run natively, outside of Docker or Kube:

    sudo $GOPATH/bin/pharos-host-upgrades

This mode only requires `systemd` DBUS access. Kube locks and configs are not supported.

#### Docker

Example `docker run` options:

    docker run --rm --name host-upgrades -v $PWD/config:/etc/host-upgrades -v /run/host-upgrades:/run/host-upgrades -v /var/run/dbus:/var/run/dbus -v /run/log/journal:/run/log/journal --privileged kontena/pharos-host-upgrades

#### Kubernetes

See the [example kube resources](./resources):

    kubectl apply -f ./resources

### CLI Options
```
Usage of pharos-host-upgrades:$ cd ^C
  -alsologtostderr
    	log to standard error as well as files
  -config-path string
    	Path to configmap dir (default "/etc/host-upgrades")
  -host-mount string
    	Path to host mount (default "/run/host-upgrades")
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

#### `--schedule`

See https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format

Variant of a standard crontab with a leading seconds field.

Examples:

* `0 15 5 * * *` - every day at 05:15:00u

## Configuration

The kube DaemonSet also supports an optional ConfigMap with configuration files for the host OS package upgrade tools. The ConfigMap should be mounted at `--config-path=/etc/host-upgrades`, and the `--host-mount=/run/host-upgrades` path should be bind-mounted from the host.

### Ubuntu `unattended-upgrades.conf`

Refer to the host `/etc/apt/apt.conf.d/50unattended-upgrades` config file shipped by the `unattended-upgrades` package.

Note overriding the `Unattended-Upgrade::Allowed-Origins` list requires the use of a `#clear` directive, to have the list items in the config override those in the default host config, rather than merging the two lists together. See the sample [`unattended-upgrades.conf`](./config/unattended-upgrades.conf) for an example:

```
// Override system /etc/apt/apt.conf.d/50unattended-upgrades
#clear "Unattended-Upgrade::Allowed-Origins";

// Automatically upgrade packages from these (origin:archive) pairs
Unattended-Upgrade::Allowed-Origins {
  ...
};
```

### CentOS `yum-cron.conf`

Refer to the host `/etc/yum/yum-cron.conf` config file shipped by the `yum-cron` package.

Note that the default `random_sleep` value will delay upgrades across the entire cluster, and should be disabled. See the sample [`yum-cron.conf`](./config/yum-cron.conf) for an example:

```
[commands]
random_sleep = 0
```

## Development

Using the vagrant machines:

    $ vagrant up ubuntu
    $ vagrant up centos-7

Note that the `centos/7` box does not support shared folders... vagrant falls back to rsync, so you must `vagrant rsync centos-7` before building the image after any edits.

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
