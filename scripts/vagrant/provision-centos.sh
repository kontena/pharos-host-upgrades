set -uex

yum -q install -y \
  yum-utils \
  yum-cron

## Docker
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum -q install -y docker-ce

test -d /etc/docker || mkdir -p /etc/docker
cat <<EOF >/etc/docker/daemon.json
{
    "storage-driver": "overlay2"
}
EOF

systemctl enable docker
systemctl start docker

usermod --append --groups=docker vagrant

## Kubernetes
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

# ensure iptables for bridged traffic
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

# disable selinux for containers
setenforce 0

# configure cgroup driver
cat <<EOF > /etc/systemd/system/kubelet.service.d/15-cgroup-args.conf
[Service]
Environment="KUBELET_CGROUP_ARGS=--cgroup-driver=cgroupfs"
EOF

# disable swap
swapoff -a
sed -i '/swap/d' /etc/fstab

yum -q install -y kubelet kubeadm kubectl

systemctl enable kubelet && systemctl start kubelet

# configure CNI
test -d /etc/cni/net.d || mkdir -p /etc/cni/net.d
cat <<EOF >/etc/cni/net.d/bridge.json
{
	"name": "cni",
	"type": "bridge",
	"bridge": "cni",
	"isDefaultGateway": true,
	"forceAddress": false,
	"ipMasq": true,
	"hairpinMode": true,
	"ipam": {
		"type": "host-local",
		"subnet": "10.10.2.0/24"
	}
}
EOF

# setup kube
test -e /etc/kubernetes/admin.conf || kubeadm init
test -e ~/.kube || mkdir ~/.kube
cat /etc/kubernetes/admin.conf > ~/.kube/config
