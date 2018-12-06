[![CircleCI](https://circleci.com/gh/logdna/logspout.svg?style=svg)](https://circleci.com/gh/logdna/logspout)

# LogDNA LogSpout

A Docker LogSpout image to stream logs from your containers to LogDNA.

## How to Use

### Environment Variables
* __LOGDNA_KEY__: LogDNA Ingestion Key, *Required*
* __FILTER_NAME__: Filter by Container Name with Wildcards, *Optional*
* __FILTER_ID__: Filter by Container ID with Wildcards, *Optional*
* __FILTER_SOURCES__: Filter by Comma-separated List of Sources, *Optional*
* __FILTER_LABELS__: Filter by Comma-separated List of Labels, *Optional*
* __HOSTNAME__: Alternative Hostname, *Optional*
* __TAGS__: Comma-separated List of Tags, *Optional*
* __LOGDNA_URL__: Specific Endpoint to Stream Log into, *Optional*

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

Append the following to your ECS cloud configuration:
```yaml
services:
  logspout:
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

* Create a LogDNA stack from the Rancher compose file below:
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

Append the following to your Docker Swarm compose file:
```yaml
version: "3"
networks:
- logging:
    services:
      logdna:
        image: logdna/logspout:latest
        volumes:
        - /var/run/docker.sock:/var/run/docker.sock
        environment:
        - LOGDNA_KEY="<LogDNA Ingestion Key>"
```

### Notes

Do not forget to add `-u root` (in CLI) or `user: root` (in YAML) in case of having permission issues.

## Contributing

Contributions are always welcome. See the [contributing guide](/CONTRIBUTING.md) to learn how you can help. Build instructions for the agent are also in the guide.
