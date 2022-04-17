# Monitor your services

Minimalistic self-hosted monitoring API with emphasis on developer experience.

## Launch container

```bash
docker network create statuspage

docker run -p 3001:3000 --network=statuspage -e GRAPHQL_URI=http://statuspage:8888/graphql --name=statuspage-ui ghcr.io/kfsoftware/statuspage-ui:sha-014d5b4

docker run -p 8888:8888 --network=statuspage --name=statuspage ghcr.io/kfsoftware/statuspage:dviejo-v0.0.1-beta server --address="127.0.0.1:8888"

```
