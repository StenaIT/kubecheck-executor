# kubecheck-executor

A tiny program that calls [kubecheck](https://github.com/StenaIT/kubecheck) over HTTP, analyzes the checks and invokes webhooks based on the results.

[![Docker Automated build](https://img.shields.io/docker/cloud/automated/stena/kubecheck-executor.svg?style=for-the-badge)](https://hub.docker.com/r/stena/kubecheck-executor/)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/stena/kubecheck-executor.svg?style=for-the-badge)](https://hub.docker.com/r/stena/kubecheck-executor/)
[![MicroBadger Size](https://img.shields.io/microbadger/image-size/stena/kubecheck-executor.svg?style=for-the-badge)](https://hub.docker.com/r/stena/kubecheck-executor/)
[![Docker Pulls](https://img.shields.io/docker/pulls/stena/kubecheck-executor.svg?style=for-the-badge)](https://hub.docker.com/r/dotnetmentor/kubecheck-executor/)

## Configuration

| Environment variable         | Description                                                                     | Required |
|------------------------------|---------------------------------------------------------------------------------|----------|
| KUBECHECK_URL                | The address where kubecheck is running                                          | true     |
| FAILED_WEBHOOK_URL           | The webhook URL to invoke when a check fails                                    | true     |
| EXIT_WEBHOOK_URL             | The webhook URL to invoke before the program exits                              | true     |
| FAILED_WEBHOOK_BODY_TEMPLATE | The webhook body template used for the "failed" webhook. Supports Go templates. | false    |
| EXIT_WEBHOOK_BODY_TEMPLATE   | The webhook body template used for the "exit" webhook. Supports Go templating.  | false    |


## Webhook Body Templates

Webhooks may have a body sent along to the webhook url by specifying a body template (`*_BODY_TEMPLATE`). The body template supports interpolation using Go templates.

Example body template:
```
{{ .Name }} - {{ .Check }} : {{ .Status }}
```

Example rendered body:
```
kubecheck-executor - http-check : failed
```

The following are the fields that can be used in each body template.

### Failed webhook

| Field       | Description                         |
|-------------|-------------------------------------|
| Name        | The name of the program             |
| Check       | The name of the check that failed   |
| Description | The description of the failed check |
| Status      | The status of the check             |
| Reason      | The reason for the failed check     |
| Input       | The input for the failed check      |
| Output      | The output of the failed check      |
| Timestamp   | The timestamp of the event          |

### Exit webhook

| Field       | Description                  |
|-------------|------------------------------|
| Name        | The name of the program      |
| Description | The description of the event |
| Timestamp   | The timestamp of the event   |
