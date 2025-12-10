tag=v0.1.1

docker build -t ghcr.io/BeaverHouse/ba-data-process:$tag .
docker push ghcr.io/BeaverHouse/ba-data-process:$tag
