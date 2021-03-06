#!/bin/sh

# compile application for the major operating systems
gox -osarch='linux/amd64' -osarch='darwin/amd64' -osarch='windows/amd64' -output='fastly.{{.OS}}'

# run the relevant compatible compiled binary for this container's OS
./fastly.linux -h
