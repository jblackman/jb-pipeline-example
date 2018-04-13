# Example concourse pipelines

This project contains worked examples demonstrating deployment pipelines from Github to Cloud Foundry

It is the source material for a presentation I made about Concourse at the Office for National Statistics.

# Using

If you want to quickly play with some Concourse pipelines, you're welcome to fork this repository (you'll
need to do this if you want your commits to trigger pipeline activity).

## Quick start pre-requisites

Install these components on your machine:

1. Docker
2. The `fly` CLI (https://concourse-ci.org/download.html)

Get yourself a Cloud Foundry account. You may already have one of these, but if you don't
you can register for a trial account at (https://run.pivotal.io).

## Set up Concourse

Go to (https://github.com/concourse/concourse-docker). Download the `docker-compose-quickstart.yml` file,
renaming it to `docker-compose.yml`. Then run:

```
docker-compose up
```

That will kick off Concourse on your local machine, running in Docker. You can kill it in the normal
docker-compose way, but be aware that when you restart, any installed pipelines will have been lost.
That said, it's so easy to add the pipeline back in again.

## Set up parameters

The [example-params.yml](ci/pipelines/example-params.yml) file shows you the parameters that the example pipelines
require. Copy and edit this file, inserting your Cloud Foundry credentials and the URLs that your 
apps will listen on. In my repo, it's configured for Pivotal Web Services.

The app URLs are defined in the manifest files (under `go-api` and `web-app` folders) as the
`Hostname:` parameters.

## Install the pipeline

Assuming your params.yml was saved to the top-level directory, it's pretty easy:

```
fly -t local login
fly -t local set-pipeline -p demo -c ci/pipelines/001.unit-test.yml -l params.yml
```

## Now, play

You can log into your Concourse on (http://localhost:8080). The Docker version doesn't require authentication.

Commit and push a change to (go-api/main.go), for example and watch the pipeline go.

Note: Concourse polls the git repo for latest commit changes, and could take 30s or so to trigger.
