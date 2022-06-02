#!/bin/bash

# ** Builds docker image called 'graph-based-server-img', automatically tagged 'latest' **

# stop on error
set -e

# run cmd
docker build -t graph-based-server-img .