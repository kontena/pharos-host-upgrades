set -uex

apt-get update
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common

curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor > /etc/apt/trusted.gpg.d/docker.gpg
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor > /etc/apt/trusted.gpg.d/kubernetes.gpg

add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/debian \
   $(lsb_release -cs) \
   stable"

cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF

apt-get update
apt-get install -y docker-ce
apt-get install -y kubelet kubeadm kubectl
apt-get install -y unattended-upgrades

cat <<EOF >/etc/docker/daemon.json
{
    "storage-driver": "overlay2"
}
EOF

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
		"subnet": "10.10.1.0/24"
	}
}
EOF

adduser vagrant docker

swapoff -a

test -e /etc/kubernetes/admin.conf || kubeadm init
test -e ~/.kube || mkdir ~/.kube
cat /etc/kubernetes/admin.conf > ~/.kube/config
