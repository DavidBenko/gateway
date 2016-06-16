#!/bin/bash

MACHINE=$(docker-machine status oracle 2> /dev/null)

if [ "$MACHINE" == "Running" ]; then
  echo 'Oracle VM is already running'
else
  if [ "$MACHINE" == "Stopped" ]; then
    echo 'Starting VM'
    docker-machine start oracle
    if [ "$?" -ne "0" ]; then
      echo "Failed to start oracle VM"
      exit 1
    fi
    docker-machine regenerate-certs oracle
    if [ "$?" -ne "0" ]; then
      echo "Failed to regenerate certificates. Please fix the existing VM and try again."
      exit 2
    fi
  else
    echo 'Creating VM'
    docker-machine create --driver=virtualbox --virtualbox-memory "2048" --virtualbox-disk-size "200000" oracle
    if [ "$?" -ne "0" ]; then
      echo "Failed to create oracle VM"
      exit 3
    fi
  fi
fi

echo 'Setting environment'
eval "$(docker-machine env oracle)"

RUNNING=$(docker inspect -f "{{.State.Running}}" orcl 2> /dev/null)

if [ "$RUNNING" == "true" ]; then
  echo 'Oracle DB is already running'
else
  echo 'Starting Oracle 12c DB on port 1521... grab a snickers'
  docker run --privileged -d -p 1521:1521 --name orcl anypresence/oracle-12c
fi
