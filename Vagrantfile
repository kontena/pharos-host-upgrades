# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Ubuntu xenial
  config.vm.define "ubuntu" do |machine|
    machine.vm.box = "ubuntu/xenial64"
    machine.vm.hostname = "ubuntu-xenial"
    machine.vm.provision "shell", path: 'scripts/vagrant/provision-ubuntu.sh'
    machine.vm.network "private_network", ip: "192.168.56.11"
  end

  # CentOS 7
  config.vm.define "centos-7" do |machine|
    machine.vm.box = "centos/7"
    machine.vm.hostname = "centos-7"
    machine.vm.provision "shell", path: 'scripts/vagrant/provision-centos.sh'
    machine.vm.network "private_network", ip: "192.168.56.12"
  end

  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  # config.vm.box_check_update = false

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine. In the example below,
  # accessing "localhost:8080" will access port 80 on the guest machine.
  # NOTE: This will enable public access to the opened port
  # config.vm.network "forwarded_port", guest: 80, host: 8080

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine and only allow access
  # via 127.0.0.1 to disable public access
  # config.vm.network "forwarded_port", guest: 80, host: 8080, host_ip: "127.0.0.1"


  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  # config.vm.network "public_network"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  # config.vm.synced_folder "../data", "/vagrant_data"

  # Provider-specific configuration so you can fine-tune various
  # backing providers for Vagrant. These expose provider-specific options.
  config.vm.provider "virtualbox" do |vb|
      # Customize the amount of memory on the VM:
      vb.memory = "2048"

      # Get rid of the annoying ./ubuntu-xenial-16.04-cloudimg-console.log file
      vb.customize [ "modifyvm", :id, "--uartmode1", "disconnected" ]
  end
  #
  # View the documentation for the provider you are using for more
  # information on available options.
end
