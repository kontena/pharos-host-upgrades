[![Build Status](https://travis-ci.com/kontena/pharos-host-upgrades.svg?token=pcXcAqCByqv2epJ6v1zo&branch=master)](https://travis-ci.com/kontena/pharos-host-upgrades)

# pharos-host-upgrades

See [`resources`](./resources) for example kube manifests.

### Supported host OS configurations

#### Ubuntu 16.04

    apt-get install unattended-upgrades

Disable `apt-periodic` `unattended-upgrades`:

```
sed -i 's/APT::Periodic::Unattended-Upgrade .*/APT::Periodic::Unattended-Upgrade "0";/' /etc/apt/apt.conf.d/20auto-upgrades
```

Alternatively, you can completely disable `apt-periodic`, as the upgrade process will also run `apt-get update` before the `unattended-upgrades`:

    systemctl stop apt-daily.timer apt-daily-upgrade.timer
    systemctl disable apt-daily.timer apt-daily-upgrade.timer

The set of allowed origins for package upgrades can either be configured via the system `/etc/apt/apt.conf.d/50unattended-upgrades` or using a [ConfigMap `unattended-upgrades.conf`](#ubuntu-unattended-upgradesconf), but the default settings are reasonable.

#### CentOS 7

    yum install yum-cron

Ensure that the `yum-cron` service is stopped and disabled, as this will interfere with `pharos-host-upgrades`:

    systemctl stop yum-cron.service
    systemctl disable yum-cron.service

The default `yum-cron.conf` `random_sleep = 360` should also be disabled, either via the default system `/etc/yum/yum-cron.conf` file, or using a [ConfigMap `yum-cron.conf`](#centos-yum-cronconf). The default `yum-cron.conf` will only download updates, and requires `apply_updates = yes` to actually upgrade the host.

## Kubernetes Integrations

When configured to run as a kube DaemonSet pod (using `KUBE_*` envs), the following kube API integrations can be used:

### DaemonSet Locking

The host upgrades will only run while holding a lock on the kube daemonset, ensuring that only one host upgrades at a time. This lock is also held during a reboot, and released once the pod restarts.

The lock is implemented as a `pharos-host-upgrades.kontena.io/lock` annotation on the DaemonSet.

### Node Draining

If configured with `--reboot --drain`, the kube node will be drained before rebooting, marking the node as unschedulable and evicting pods to move them to other nodes for the duration of the reboot.

The node will be uncordoned once the `host-upgrades` pod is restarted, but only if the `pharos-host-upgrades.kontena.io/drain` annotation was previously set as a result a drain + reboot triggered by the `host-upgrades` pod. If the pod is restarted while the node was otherwise drained, it will not be uncordoned.

### Node Conditions

The kube node `.Status.Conditions` will be updated based on the result of the host upgrades:

#### `HostUpgrades`

The `HostUpgrades` condition describes the state of the upgrade command itself.

The condition will be `True` if the upgrade was run and the node is now up to date, with a message describing any packages upgraded during the last run.

 The condition will be `False` if the host is not up to date. This will happen in the `RebootRequired` case, where the host requires a reboot to finish applying the upgrades.

In case of the upgrade failing, the condition will be `Unknown`, with a message describing the error.

#### `HostUpgradesReboot`

The `HostUpgradesReboot` condition will be `True` if the host requires a reboot to finish applying upgrades, and `False` otherwise.

### Supported Kube Versions

 * Kubernetes 1.10

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

The DaemonSet configures the `KUBE_*` envs, enabling the kubernets integrations.

### CLI Options
```
Usage of pharos-host-upgrades:
  -alsologtostderr
    	log to standard error as well as files
  -config-path string
    	Path to configmap dir (default "/etc/host-upgrades")
  -drain
    	Drain kube node before reboot, uncordon after reboot
  -host-mount string
    	Path to shared mount with host. Must be under /run to reset when rebooting! (default "/run/host-upgrades")
  -kube-daemonset string
    	Name of kube DaemonSet (KUBE_DAEMONSET) (default "host-upgrades")
  -kube-namespace string
    	Name of kube Namespace (KUBE_NAMESPACE) (default "kube-system")
  -kube-node string
    	Name of kube Node (KUBE_NODE) (default "ubuntu-xenial")
  -log_backtrace_at value
    	when logging hits line file:N, emit a stack trace
  -log_dir string
    	If non-empty, write log files in this directory
  -logtostderr
    	log to standard error instead of files
  -reboot
    	Reboot if required
  -reboot-timeout duration
    	Wait for system to shutdown when rebooting (default 5m0s)
  -schedule string
    	Scheduled upgrade (cron syntax)
  -stderrthreshold value
    	logs at or above this threshold go to stderr
  -v value
    	log level for V logs
  -vmodule value
    	comma-separated list of pattern=N settings for file-filtered logging
command terminated with exit code 2
```

#### `--schedule`

See https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format (no seconds field).

Standard crontab with five fields (Minutes, Hours, Day of month, Month, Day of week).

Examples:

* `15 5 * * *` - every day at 05:15
* `1 0 * * SUN` - every sunday at 01:00
* `@daily` at midnight

#### `--reboot` `--reboot-timeout=...`

Reboot the host after upgrades, if required.

#### `--drain`

Drain the kube node before rebooting, and uncordon once restarted.

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
