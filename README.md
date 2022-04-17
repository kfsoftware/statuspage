# Monitor your services

Minimalistic self-hosted monitoring API with emphasis on developer experience.

## Launch container

```bash
docker network create statuspage

docker run -p 3001:3000 --network=statuspage -e GRAPHQL_URI=http://statuspage:8888/graphql --name=statuspage-ui ghcr.io/kfsoftware/statuspage-ui:sha-014d5b4



```
