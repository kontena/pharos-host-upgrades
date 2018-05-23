set -uex

yum -q install -y \
  yum-utils \
  yum-cron

yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum -q install -y docker-ce

systemctl start docker

usermod --append --groups=docker vagrant
