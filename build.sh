#!/bin/bash

docker build . -t kairos-re-unlock:latest
docker compose up aurora-boot
mv ./build/*.iso ./build/boot.iso
rm -rf ./qemu
docker compose up qemu
