#!/bin/bash

echo 'provisioning apache2'

export DEBIAN_FRONTEND=noninteractive

if [ -f $HOME/.bashrc ]; then
  echo "export DEBIAN_FRONTEND=noninteractive" | sudo tee -a $HOME/.bashrc
fi

sudo apt-get update
sudo apt-get -y install apache2
echo '<!doctype html><html><body><h1>Hello from stackql droplet auto-provisioned.</h1></body></html>' | sudo tee /var/www/html/index.html

echo 'provisioning complete'

