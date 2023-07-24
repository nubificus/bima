# Sample Unikernel images

## Download latest bima

```bash
curl -L -o bima https://s3.nubificus.co.uk/nbfc-assets/github/bima/dist/v0.2.0/$(uname -m)/bima_$(uname -m)

```

## Create a new image with bima

```bash
BUILD_DIR=$(mktemp -d)
pushd $BUILD_DIR
curl -L -o redis.hvt https://s3.nubificus.co.uk/bima/redis.hvt
curl -L -o redis.conf https://s3.nubificus.co.uk/bima/redis.conf

tee Containerfile > /dev/null <<EOT
FROM scratch
COPY redis.hvt /unikernels/redis.hvt
COPY redis.conf /conf/redis.conf

LABEL com.urunc.unikernel.binary=/unikernel/redis.hvt
LABEL "com.urunc.unikernel.cmdline"='redis-server /data/conf/redis.conf'
LABEL "com.urunc.unikernel.unikernelType"="rumprun"
LABEL "com.urunc.unikernel.hypervisor"="hvt"
EOT

sudo bima build -t harbor.nbfc.io/nubificus/redis-hvt:latest -f Containerfile .
popd
```

We can verify the image was successfully created:

```bash
sudo ctr image ls
```

We can push the image to Docker Hub or any other registry:

```bash
sudo ctr image push harbor.nbfc.io/nubificus/redis-hvt:latest -u '$USERNAME:$PASSWORD'
```