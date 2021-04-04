# -*- mode: ruby -*-
# vi: set ft=ruby :

# This is a just a networking test
Vagrant.configure("2") do |config|
  config.vm.define "frontend" do |frontend|
    frontend.vm.box = "hashicorp/bionic64"
    frontend.vm.network "private_network", ip: "192.168.50.2"
    frontend.vm.provider "virtualbox" do |v|
      v.memory = 512 # docker requires at least 512Mb
      v.cpus = 1
    end
  end

  config.vm.define :api do |api|
    api.vm.hostname = "api"
    api.vm.box = "hashicorp/bionic64"
    api.vm.network "private_network", ip: "192.168.50.3"
    api.vm.provider "virtualbox" do |v|
      v.memory = 512 # docker requires at least 512Mb
      v.cpus = 1
    end
  end

  config.vm.define "neo4j" do |neo4j|
    neo4j.vm.box = "hashicorp/bionic64"
    neo4j.vm.network "private_network", ip: "192.168.50.4"
    neo4j.vm.provider "virtualbox" do |v|
      v.memory = 512
      v.cpus = 1
    end  
  end

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook.yml"
  end
end
