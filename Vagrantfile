# -*- mode: ruby -*-
# vi: set ft=ruby :

# This is a just a networking test
Vagrant.configure("2") do |config|
  config.vm.define :api do |api|
    api.vm.hostname = "api"
    api.vm.box = "hashicorp/bionic64"
    api.vm.network "private_network", ip: "192.168.50.6"
    api.vm.provider "virtualbox" do |v|
      v.memory = 512
      v.cpus = 1
    end
  end

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook_api.yml"
  end
end
