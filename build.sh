tag=v0.1.2

docker build -t ghcr.io/beaverhouse/ba-data-process:$tag .
docker push ghcr.io/beaverhouse/ba-data-process:$tag
