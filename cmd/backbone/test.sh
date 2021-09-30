#!/bin/sh 

# Declaring routes

./backbone \
    --api pull image api:///pull/layers \
    --api pull layers api:///pull/manifest \
    --api pull manifest api:///v2/manifests \
    --api v2 manifests api:///v2/blobs \
    --api v2 blobs https://

# Declaring a reference

./backbone \
    --api pull image api:///pull/layers \
    --api pull layers api:///pull/manifest \
    --api pull manifest api:///v2/manifests \
    --api v2 manifests api:///v2/blobs \
    --api v2 blobs https:// \
    --reference ref://latest@registry-1.docker.io/library/ubuntu

# Declaring nodes in the control group

./backbone \
    --node api://pull/image \
    --node api://pull/layers \

# Nodes can arrange themselves 

