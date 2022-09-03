#! /bin/bash
set -e
DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

DATA_DIR=data-1
(docker stop postgres && docker rm postgres) || true
sudo rm -rf $DATA_DIR

docker run \
    -d \
    -p 5432:5432 \
    --name postgres \
    -e POSTGRES_PASSWORD=P@ssw0rd \
    -e PGDATA=/var/lib/postgresql/data \
    -v "$DIR/$DATA_DIR":/var/lib/postgresql/data \
    -v "$DIR/init-1":/docker-entrypoint-initdb.d \
    postgres:13.4

