[![CircleCI](https://circleci.com/gh/logdna/logspout.svg?style=svg)](https://circleci.com/gh/logdna/logspout)

# LogDNA LogSpout

A Docker LogSpout image to stream logs from your containers to LogDNA.

## Change Log

### v1.2.0 - Released on May 29, 2019

* Added Tagged Build;
* Added [Semantic Versioning](http://semver.org);
* Added [`CHANGELOG.md`](https://github.com/logdna/logspout/blob/master/CHANGELOG.md);
* Updated [`LICENSE`](https://github.com/logdna/logspout/blob/master/LICENSE);
* Enriched the Adapter Configuration;
* Added 12 New Environment Variable Options;
* Implemented Retry Mechanism;
* Added Message Sanitization;
* Added Capturing `m.Container.State.Pid`;
* Changed Buffer Limit from the Length to the Byte Size;
* Polished Some Debug Statements.

### v1.1.0

* Getting Tags from Templates

### v1.0.0

* Initial Release

## How to Use

### Environment Variables

The following variables can be used to tune the `logspout` for the specific use case.

### Log Router Specific

The following variables can be used for filtering the logs streaming into `LogSpout`. More information can be found [here](https://github.com/gliderlabs/logspout/tree/0da75a223db992cd5abc836796174588ddfc62b4/routesapi#routes-resource).

* __FILTER_NAME__:
  * __Description__: Filter by Container Name with Wildcards
  * __Required__:    No

* __FILTER_ID__:
  * __Description__: Filter by Container ID with Wildcards
  * __Required__:    No

* __FILTER_SOURCES__:
  * __Description__: Filter by Comma-Separated List of Sources
  * __Required__:    No

* __FILTER_LABELS__:
  * __Description__: Filter by Comma-Separated List of Labels
  * __Required__:    No

### Ingestion Specific

The following variables can be used for customizing the payloads `LogSpout` sends to `LogDNA`.

* __LOGDNA_KEY__: LogDNA Ingestion Key, *Required*
* __HOSTNAME__: Alternative Hostname, *Optional*
* __TAGS__: Comma-Separated List of Tags, *Optional*
* __LOGDNA_URL__: Specific Endpoint to Stream Log into, *Optional*, *
* __VERBOSE__: 

### Related to HTTP Client
* __DIAL_KEEP_ALIVE__:
* __DIAL_TIMEOUT__:
* __EXPECT_CONTINUE_TIMEOUT__:
* __IDLE_CONN_TIMEOUT__:
* __HTTP_CLIENT_TIMEOUT__:
* __TLS_HANDSHAKE_TIMEOUT__:

### Limits
* __FLUSH_INTERVAL__:
* __INACTIVITY_TIMEOUT__:
* __MAX_BUFFER_SIZE__:
* __MAX_LINE_LENGTH__:
* __MAX_REQUEST_RETRY__:

### Docker

Create and run container named *logdna* from this image using CLI:
```bash
sudo docker run --name="logdna" --restart=always \
-d -v=/var/run/docker.sock:/var/run/docker.sock \
-e LOGDNA_KEY="<LogDNA Ingestion Key>" \
logdna/logspout:latest
```

### Docker Cloud

Append the following to your Docker Cloud stackfile:
```yaml
logdna:
  autoredeploy: true
  deployment_strategy: every_node
  environment:
    - LOGDNA_KEY="<LogDNA Ingestion Key>"
    - TAGS='{{.Container.Config.Hostname}}'
  image: 'logdna/logspout:latest'
  restart: always
  volumes:
    - '/var/run/docker.sock:/var/run/docker.sock'
```

### Elastic Container Service (ECS)

Modify your ECS Cloud Configuration file to have `LogDNA` Service as described below:
```yaml
services:
  logdna:
    environment:
        - LOGDNA_KEY="<LogDNA Ingestion Key>"
        - TAGS='{{ if .Container.Config.Labels }}{{index .Container.Config.Labels "com.amazonaws.ecs.task-definition-family"}}:{{index .Container.Config.Labels "com.amazonaws.ecs.container-name"}}{{ else }}{{.ContainerName}}{{ end }}'
    image: logdna/logspout:latest
    restart: always
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      mode: global
```

### Rancher

Modify your Rancher Compose Stackfile to have `LogDNA` Service as described below:
```yaml
version: '2'
services:
  logdna:
    image: logdna/logspout:latest
    environment:
      LOGDNA_KEY="<LogDNA Ingestion Key>"
    restart: always
    labels:
      io.rancher.container.hostname_override: container_name
      io.rancher.container.pull_image: always
      io.rancher.os.scope: system
    volumes:
    - /var/run/docker.sock:/tmp/docker.sock
```

### Docker Swarm

Modify your Docker Swarm Compose file to have `LogDNA` Service as described below:
```yaml
version: "3"
networks:
  logging:
services:
  logdna:
    image: logdna/logspout:latest
    networks:
      - logging
    volumes:
      - /etc/hostname:/etc/host_hostname:ro
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - LOGDNA_KEY="<LogDNA Ingestion Key>"
    deploy:
      mode: global
```

### Notes

Do not forget to add `-u root` (in CLI) or `user: root` (in YAML) in case of having permission issues.

## Contributing

Contributions are always welcome. See the [contributing guide](/CONTRIBUTING.md) to learn how you can help. Build instructions for the agent are also in the guide.
