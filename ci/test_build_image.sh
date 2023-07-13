#!/bin/bash

# This script performs a very simple functionality test
echo "Running basic functionality tests..."
echo ""
# Create a sample Containerfile
tee Tempfile > /dev/null <<EOT
FROM scratch
ARCH amd64
COPY VERSION /unikernel/test-redis.hvt
COPY cmd /data/
LABEL "com.urunc.unikernel.cmdline"="{"cmdline":"redis-server /data/conf/redis.conf","net":{"if":"ukvmif0","cloner":"True","type":"inet","method":"static","addr":"10.0.66.2","mask":"24","gw":"10.0.66.1"},"blk":{"source":"etfs","path":"/dev/ld0a","fstype":"blk","mountpoint":"/data"}}"
LABEL "com.urunc.unikernel.unikernelType"="rumprun"
LABEL "com.urunc.unikernel.hypervisor"="qemu"
LABEL com.urunc.unikernel.binary=/unikernel/test-redis.hvt
EOT

# Change that to get better logs
export BIMA_LOG=fatal

IMAGE_PREFIX=bima
IMAGE_NAME=testimage
IMAGE_TAG=$(git describe --dirty --long --always)

function fatal {
    echo "$1"
    rm Tempfile
    exit 1
}

# Build an image and export to tar
bima build -t $IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG --tar -f Tempfile .
FOUND=$(ls | grep "$IMAGE_NAME" | grep "$IMAGE_TAG")
if [ ! -n "$FOUND" ]
then
	fatal "Image tarball not found"
else
    echo "Build tarball:                              OK"
fi
rm $IMAGE_NAME:$IMAGE_TAG

# Build an image and import to ctr
bima build -t $IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG -f Tempfile .
FOUND=$(ctr image ls  -q | grep "$IMAGE_NAME" | grep "$IMAGE_TAG")
if [ ! -n "$FOUND" ]
then
	fatal "Image not imported in containerd"
else
    echo "Import image in containerd:                 OK"
fi
ctr image rm docker.io/$IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG  > /dev/null

# Build an image with custom namespace and import to ctr
bima build -t $IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG -n testns -f Tempfile .
FOUND=$(ctr -n testns image ls  -q | grep "$IMAGE_NAME" | grep "$IMAGE_TAG")
if [ ! -n "$FOUND" ]
then
	fatal "Image not imported in namespace"
else
    echo "Import image in namespace:                  OK"
fi
ctr -n testns image rm docker.io/$IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG  > /dev/null


# Build an image with custom snapshotter and import to ctr
bima build -t $IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG --snapshotter overlayfs -f Tempfile .
FOUND=$(ctr image ls  -q | grep "$IMAGE_NAME" | grep "$IMAGE_TAG")
if [ ! -n "$FOUND" ]
then
	fatal "Image not imported in namespace"
else
    echo "Import image  with custom snapshotter:      OK"
fi
ctr image rm docker.io/$IMAGE_PREFIX/$IMAGE_NAME:$IMAGE_TAG > /dev/null

