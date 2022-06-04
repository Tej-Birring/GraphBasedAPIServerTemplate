#!/bin/bash

# ** Deploy docker image to 'Container Registry' of the Heroku project. **
# ** Make sure to do `$heroku login` and `$heroku container:login` first. **

# Stop on error
set -e

# Run cmd: push image to registry
# (Builds the image from the Dockerfile and pushes the image to Heroku Container Registry for this Heroku project.)
heroku container:push web --app graph-based-server

# Run cmd: release image to app
# (Makes the image 'go live' by initiating a container instance for this Heroku project.)
# See project settings in `dashboard.heroku.com` to set environment variables (called "Config Vars").
heroku container:release web --app graph-based-server