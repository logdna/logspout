# LogDNA LogSpout
A Docker LogSpout image to stream logs from your containers to LogDNA.

## How to Use

### Environment Variables
The following variables can be used to tune `LogSpout` for specific use cases.

#### Log Router Specific
* __FILTER_NAME__: Filter by Container Name with Wildcards, *Optional*
* __FILTER_ID__: Filter by Container ID with Wildcards, *Optional*
* __FILTER_SOURCES__: Filter by Comma-Separated List of Sources, *Optional*
* __FILTER_LABELS__: Filter by Comma-Separated List of Labels, *Optional*

__Note__: More information can be found [here](https://github.com/gliderlabs/logspout/tree/0da75a223db992cd5abc836796174588ddfc62b4/routesapi#routes-resource).

#### Ingestion Specific
* __LOGDNA_KEY__: LogDNA Ingestion Key, *Required*
* __HOSTNAME__: Alternative Hostname, *Optional*
  * __Default__: System's Hostname
* __LOGDNA_URL__: Specific Endpoint to Stream Log into, *Optional*
  * __Default__: `logs.logdna.com/logs/ingest`
* __TAGS__: Comma-Separated List of Tags, *Optional*

__Note__: Logging the `LogSpout` Container is recommended to keep track of HTTP Request Errors or Exceptions.

#### Limits
* __FLUSH_INTERVAL__: How frequently batches of logs are sent (in `milliseconds`), *Optional*
  * __Default__: 250
* __HTTP_CLIENT_TIMEOUT__: Time limit (in `seconds`) for requests made by this HTTP Client, *Optional*
  * __Default__: 30
  * __Source__: [net/http/client.go#Timeout](https://github.com/golang/go/blob/master/src/net/http/client.go#L89-L104)
* __INACTIVITY_TIMEOUT__: How long to wait for inactivity before declaring failure in the `Docker API` and restarting, *Optional*
  * __Default__: 1m
  * __Note__: More information about the possible values can be found [here](https://github.com/gliderlabs/logspout#detecting-timeouts-in-docker-log-streams). Also see [`time.ParseDuration`](https://golang.org/pkg/time/#ParseDuration) for valid format as recommended [here](https://github.com/gliderlabs/logspout/blob/e671009d9df10e8139f6a4bea8adc9c7878ff4e9/router/pump.go#L112-L116).
* __MAX_BUFFER_SIZE__: The maximum size (in `mb`) of batches to ship to `LogDNA`, *Optional*
  * __Default__: 2
* __MAX_REQUEST_RETRY__: The maximum number of retries for sending a line when there are network failures, *Optional*
  * __Default__: 5

### Docker
Create and run container named *logdna* from this image using CLI:
```bash
sudo docker run --name=logdna --restart=always \
-d -v=/var/run/docker.sock:/var/run/docker.sock \
-e LOGDNA_KEY=<LogDNA Ingestion Key> \
logdna/logspout:latest
```

### Docker Cloud
Append the following to your Docker Cloud stackfile:
```yaml
logdna:
  autoredeploy: true
  deployment_strategy: every_node
  environment:
    - LOGDNA_KEY=<LogDNA Ingestion Key>
    - TAGS='{{.Container.Config.Hostname}}'
  image: logdna/logspout:latest
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
        - LOGDNA_KEY=<LogDNA Ingestion Key>
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
      LOGDNA_KEY=<LogDNA Ingestion Key>
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
      - LOGDNA_KEY=<LogDNA Ingestion Key>
    deploy:
      mode: global
```

### Notes
Do not forget to add `-u root` (in CLI) or `user: root` (in YAML) in case of having permission issues.

## Contributing
Contributions are always welcome. See the [contributing guide](/CONTRIBUTING.md) to learn how you can help. Build instructions for the agent are also in the guide.
