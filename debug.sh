#!/bin/bash

# ** Runs docker image called 'graph-based-server-img' as new container **
# ** This is 'debug' variant, so maps to port 8000 of the host machine **

# stop on error
set -e

# run cmd
docker run -p 8000:80 graph-based-server-img