#!/bin/bash

set -eux

cd jb-pipeline-example/web-app
bundle install
bundle exec rspec