# Creating a Cloud Foundry CI pipeline

As an AK engineer, you will be called on to create deployment pipelines for your
new projects. With experience, you will find that they follow a predictable flow:

1. Git push trigger
2. Unit test the component that has changed
3. Deploy to a staging environment
4. Perform acceptance tests in staging
5. Deploy to a production environment
6. Perform acceptance tests in production

Depending on the project, you can expect to add additional jobs, and how each
job is performed may vary slightly. The core as described should be there,
though.

Note. Creds/keys can be found in 1Password

## Example pipeline

This repository gives you a reasonable template for a standardized Armakuni CD
pipeline. It demonstrates a typical deployment lifecycle from git commit through
unit testing, staging deployment, staging tests and eventually to production
deployment and production acceptance tests.

In the interests of brevity we do cut some corners - for example we are basing
everything on just one repository. In a real-world application you should have a
repository for each micro-service.

We have put the pipeline into the AK Concourse instance, so it's pretty easy to
get to and push to. The website is at [https://ci.armakuni.co.uk](https://ci.armakuni.co.uk/teams/ak-example/pipelines/ci-pipeline-example), and you should
choose the `ak-example` team.

To push the pipeline:

```
fly -t ak login -c https://ci.armakuni.co.uk -u armakuni -p {password} -n ak-example
fly -t ak set-pipeline -p ci-pipeline-example -c ci/pipeline.yaml -l ci/params.yaml
```

## Useful links

* [Pipeline](https://ci.armakuni.co.uk/teams/main/pipelines/ci-pipeline-example)
* [Staging API](https://ak-example-staging-go-api-ak.cfapps.io/)
* [Production API](https://ak-example-production-go-api-ak.cfapps.io/)
* [Staging Web](https://ak-example-staging-web-app-ak.cfapps.io/)
* [Production Web](https://ak-example-production-web-app-ak.cfapps.io/)

## Project repository organisation

In a real project, you will of course separate the applications into separate
repositories. You should also create a "bootstrap" repository in which you will
place tools to stand up the whole project, including the CI pipelines and tasks.

As the pipeline will need to access the git repositories programatically, you
should generate a public/private key pair and install the public key into the
project, or into each repository. In most cases, access using this key will be
read-only.

## Triggering from Git

The `git` resource type will poll the specified resource and trigger on a push.
As your applications are all configured as microservices, you wish to unit test
each one independently, which means that you should create a separate job for
each application. This example shows a Go microservice, where you have created
a unit test task.

```yaml
resources:
- name: bootstrap
  type: git
  source:
    uri: git@bitbucket.org/armakuni/my-project/bootstrap.git
    branch: master
    private_key: {{git_private_key}}

- name: my-microservice
  type: git
  source:
    uri: git@bitbucket.org/armakuni/my-project/my-microservice.git
    branch: master
    private_key: {{git_private_key}}
...
jobs:
- name: my-microservice-unit
  plan:
  - get: bootstrap
  - get: my-microservice
    trigger: true
  - task: unit_test
    file: bootstrap/ci/tasks/unit_test_go.yml
    params:
      PROJECT: my-microservice
```

## Deploying to staging or production

Deployment tasks depend on the operating environment. Armakuni is pretty closely
involved with Cloud Foundry, which has the benefit that deployments subsume any
compilation step and there are no artifacts to worry about. The example pipeline
shows how this would be done, using the `cf` resource.

```yaml
resources:
...
- name: cf-staging
  type: cf
  source:
    api: {{cf_api}}
    username: {{cf_username}}
    password: {{cf_password}}
    organization: {{cf_org}}
    space: {{cf_space_staging}}

jobs:
- name: my-microservice-staging
  plan:
  - get: bootstrap
  - get: my-microservice
    trigger: true
    passed: [ my-microservice-unit ]
  - put: cf-staging
    params:
      manifest: my-microservice/ci/manifest.yml
      path: my-microservice
```

## Acceptance tests

Once the deployment has completed, you should have acceptance tests to validate
the site. For production, these should be fairly non-invasive. In the example
pipeline, we're just querying the API endpoint. Your tests should be quite a
bit more comprehensive, including user journey tests (e.g. login, request page,
post form, validate results are OK).

## Other concepts

### Grouping deployments and tests

You can either deploy your applications to staging or production individually,
or do them all (in parallel!) in a single job. The first approach is faster and
limits the impact of your change. However, if it introduces an incompatibility
with the other applications, then the whole project will be broken until the
other applications are also deployed.

You can mitigate this risk in a couple of ways: [1] use the individual strategy for
staging deploys and the grouped strategy for production, or [2] version your
APIs so that you retain backwards compatibility, at least for until all applications
are running with the new version.

All of the acceptance tests should be run when any component changes, as you
should be regression-testing the entire environment to catch synergistic
problems. You may wish to group your acceptance tests into functional units,
however (e.g. by payment provider), so that the pipeline shows the general
area where a test failed, rather than just saying "it failed" and requiring you
to drill down every time.

### Manifest generation step

If you wish to add environment variables to your application, the `cf` resource has
a param for that, which you can specify in the `put`:

```yaml
jobs:
- name: cf-staging
  plan:
  - get: my-microservice
  - put: cf-staging
    params:
      manifest: my-microservice/ci/manifest.yml
      environment_variables:
        SHARD_ID: {{app_shard_id_staging}}
```

However, if you do need to modify the manifest, then put that in a task defining
an output file and use the modified file in the `put` step.

### Production smoke testing

Smoke tests should run regularly to validate the production platform. Their
place may not necessarily be on the deployment pipeline - it really depends
on how cluttered the pipeline already is. You can use the built in `time`
resource type to schedule the tests, or you could try the (to my mind, more
intuitive) cron resource at 
[https://github.com/pivotal-cf-experimental/cron-resource](https://github.com/pivotal-cf-experimental/cron-resource).

### Alerting on failures

If critical failures occur, such as production acceptance tests, then it is a
very good idea to alert. You can use the `on_failure` tag on both tasks and
puts to send emails and/or Slack messages. These are community resource types
linked in [Concourse pages](https://concourse-ci.org/resource-types.html).

The email resource type is a bit annoying - you have to put subject and body
into files rather than being able to specify them as strings in the pipeline.

### Adding a build phase to your pipeline

If you are deploying to a platform that does not do the compilation step for you,
you will need to add a build phase to the pipeline. This involves compiling the
software and generating an artifact (e.g. a Go binary). The output should be
storing the artifact in an object store (not Git!). The object store is then an
input resource for the deployment pipeline job.

Artifact versioning may come into the picture at this point. You need some way of
tying the particular binary version to the code - perhaps by amending the binary
name with the commit ref or using tags. Go has a cool feature whereby you can
inject the tag into the binary when compiling:

```go
package main

import "fmt"

var version string

func main() {
	fmt.Println("Version:", version)
}
```

```bash
$ go run -ldflags="-X main.version=1.2.3" main.go
Version: 1.2.3
```

### Running the pipeline without internet access

Just a few things to consider:
- if you can whitelist GitHub and Docker Hub through a proxy, do it
- otherwise, you will need to set up a local Docker registry and possibly mirror
any Github dependencies
- consider how any alerts will be propagated

.
