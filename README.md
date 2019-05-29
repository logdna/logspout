[![CircleCI](https://circleci.com/gh/logdna/logspout.svg?style=svg)](https://circleci.com/gh/logdna/logspout)

# LogDNA LogSpout

A Docker LogSpout image to stream logs from your containers to LogDNA.

## Change Log

### v1.2.0 - Released on May 29, 2019

* Added Tagged Build;
* Added [`Semantic Versioning`](http://semver.org);
* Added [`CHANGELOG.md`](https://github.com/logdna/logspout/blob/master/CHANGELOG.md);
* Updated [`LICENSE`](https://github.com/logdna/logspout/blob/master/LICENSE);
* Enriched the Adapter Configuration;
* Added 11 New Environment Variable Options;
* Implemented `Retry` Mechanism;
* Added Message Sanitization;
* Added Capturing `m.Container.State.Pid`;
* Changed Buffer Limit from the Length to the `Byte Size`;
* Polished Some Debug Statements.

### v1.1.0 - Released on December 06, 2018

* Getting Tags from Templates

### v1.0.0 - Released on August 29, 2018

* Initial Release

## How to Use

### Environment Variables

The following variables can be used to tune the `LogSpout` for the specific use case.

#### Log Router Specific

* __FILTER_NAME__: Filter by Container Name with Wildcards, *Optional*
* __FILTER_ID__: Filter by Container ID with Wildcards, *Optional*
* __FILTER_SOURCES__: Filter by Comma-Separated List of Sources, *Optional*
* __FILTER_LABELS__: Filter by Comma-Separated List of Labels, *Optional*

__Note__: More information can be found [here](https://github.com/gliderlabs/logspout/tree/0da75a223db992cd5abc836796174588ddfc62b4/routesapi#routes-resource).

#### Ingestion Specific

* __LOGDNA_KEY__: LogDNA Ingestion Key, *Required*
* __HOSTNAME__: Alternative Hostname, *Optional*
* __TAGS__: Comma-Separated List of Tags, *Optional*
* __LOGDNA_URL__: Specific Endpoint to Stream Log into, *Optional*
  * __Default__: `logs.logdna.com/logs/ingest`
* __VERBOSE__: Enabling or Disabling to Log `LogSpout` Container, *Optional*
  * __Default__: Enabled
  * __Note__: `0` to Disable

__Note__: Logging `LogSpout` Container is recommended to keep track of the HTTP Request Error or Exceptions.

#### Related to HTTP Client
* __DIAL_KEEP_ALIVE__: The interval between keep-alive probes for an active network connection.
  * __Default__: 60
  * __Source__: [net/dial.go#Timeout](https://github.com/golang/go/blob/master/src/net/dial.go#L72-L79)
* __DIAL_TIMEOUT__: The maximum amount of time (in `seconds`) a dial will wait for a connect to complete.
  * __Default__: 60
  * __Source__: [net/dial.go#Timeout](https://github.com/golang/go/blob/master/src/net/dial.go#L27-L39)
* __IDLE_CONN_TIMEOUT__: The maximum amount of time (in `seconds`) an idle (keep-alive) connection will remain idle before closing itself.
  * __Default__: 60
  * __Source__: [net/http/transport.go#IdleConnTimeout](https://github.com/golang/go/blob/master/src/net/http/transport.go#L213-L217)
* __HTTP_CLIENT_TIMEOUT__: Time limit (in `seconds`) for requests made by this HTTP Client, *Optional*
  * __Default__: 60
  * __Source__: [net/http/client.go#Timeout](https://github.com/golang/go/blob/master/src/net/http/client.go#L89-L104)
* __TLS_HANDSHAKE_TIMEOUT__: The maximum amount of time (in `seconds`) waiting to wait for a TLS handshake, *Optional*
  * __Default__: 30
  * __Source__: [net/http/transport.go#TLSHandshakeTimeout](https://github.com/golang/go/blob/master/src/net/http/transport.go#L171-L173)

#### Limits
* __FLUSH_INTERVAL__: The frequency (in `milliseconds`) to ship the batch of logs, *Optional*
  * __Default__: 250
* __INACTIVITY_TIMEOUT__: The timeout to detect failure in `Docker API` and restart, *Optional*
  * __Default__: 1m
  * __Note__: More information about the possible values can be found [here](https://github.com/gliderlabs/logspout#detecting-timeouts-in-docker-log-streams).
* __MAX_BUFFER_SIZE__: The maximum byte size (in `mb`) of the batch to ship to `LogDNA`, *Optional*
  * __Default__: 2
* __MAX_LINE_LENGTH__: The maximum character length for each line, *Optional*
  * __Default__: 16000
* __MAX_REQUEST_RETRY__: The maximum # of retries in case of network issue, *Optional*
  * __Default__: 10

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
