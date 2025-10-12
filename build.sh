#!/bin/bash
set -e

rm -rf ./qemu
rm -rf ./build

docker compose down -v
docker build . -t kairos-re-unlock:latest
docker compose up aurora-boot
mv ./build/*.iso ./build/boot.iso
docker compose up qemu
