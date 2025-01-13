#!/bin/bash

echo 'provisioning all in one stackql dashboard VM'

sudo DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get update
sudo DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get install -yq postgresql postgresql-contrib

mkdir -p /opt/stackql

curl -L https://bit.ly/stackql-zip -O && unzip stackql-zip && cp stackql /opt/stackql/ && chmod +x /opt/stackql/*

export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a
export PATH=$PATH:/opt/stackql

if [ -f $HOME/.bashrc ]; then
  echo "export DEBIAN_FRONTEND=noninteractive" | sudo tee -a $HOME/.bashrc
  echo "export NEEDRESTART_MODE=a"             | sudo tee -a $HOME/.bashrc
  echo "export PATH=$PATH:/opt/stackql"        | sudo tee -a $HOME/.bashrc
fi

echo 'provisioning complete'

