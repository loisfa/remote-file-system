# Remote File System [golang, svelte.js]

## Requirements
- GoLang
- NodesJS 10+
- Python 3
- pipenv

## Start the app
### API
Inside /api: ```go run main.go```
Launches a golang server on http://localhost:8080

NB: 
- main.go is a (long) single file for development reasons, could not develop properly on VSCOde with multiple golang files.
- Don't touch the initial files (file1.txt + file.txt) inside /api/temp-files.

### Front-end
Inside /front: ```npm run dev```
Launches a node server on http://localhost:5000. You can access the wep app at this location.

## Start python integration tests
Inside /api: ```python3 integration_tests.tests```


## Infrastructure (in progress)

### Requirements
- Virtualbox 6.1
- Vagrant 2.2+ (+sudo apt install libarchive-tools, otherwise might gett he following error on ```vagrant up```: "bsdtar not found...")

### Overview
Components:
- Frontend deployed on 1 VM (dockerized app) => Public network access (Future = gateway component for the front and the back)
- Rest API deployed on 1 VM (dockerized app): Stateless API + Stateful Filesystem (To be moved in dedicated VMs later) => Public network access (Future = gateway component for the front and the back)
- Neo4J database deployed on 1 VM (NON-dockerized app)

###
Useful commands:
vagrant init hashicorp/bionic64 => creates the Vagrantfile
vagrant up => start the machines defined in the Vagrantfile (and download them if not done yet. Around 300Mb!)
vagrant ssh [name of a machine] => ssh into a machine [name]

### TODO
1. vagrant + virtualbox + ansible => deploy frontend it its own VM and expose it the traffic

