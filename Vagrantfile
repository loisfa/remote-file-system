# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.define "neo4j" do |frontend|
    config.vm.box = "hashicorp/bionic64"
    config.vm.network "private_network", ip: "192.168.50.5"
    config.vm.provider "virtualbox" do |v|
      v.memory = 1024
      v.cpus = 1
    end  
  end

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook_neo4j.yml"
  end
end
