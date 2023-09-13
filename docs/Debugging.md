# Debugging processes

This document assumes you have followed the [Sample_Images.md](/docs/Sample_Images.md) tutorial.

```bash
BUILD_DIR=$(mktemp -d)
pushd $BUILD_DIR
sudo nerdctl run --rm -ti -v $(pwd):/workspace --privileged gntouts/ocitools:latest
```

Inside the `ocitools` container:

```bash
skopeo copy docker://harbor.nbfc.io/nubificus/redis-hvt:latest oci:redis-hvt:latest
umoci unpack --image redis-hvt:latest bundle
```

To list the annotations in the `config.json` file:

```bash
cat bundle/config.json | jq '{annotations}'
```

To list the annotations in the `urunc.json` file:

```bash
cat bundle/rootfs/urunc.json | jq
```

To add config options or annotations:

```bash
umoci config --author="Nubificus LTD" --image redis-hvt:latest
umoci config --config.label="com.urunc.unikernel.hypervisor=hvt" --image redis-hvt:latest
```

For example, to add the Labels defined in  [Sample_Images.md](/docs/Sample_Images.md) tutorial, we can:

```bash
unikernel_binary=$(echo '/unikernel/redis.hvt' | base64)
unikernel_cmdline=$(echo 'redis-server /data/conf/redis.conf' | base64)
unikernel_type=$(echo 'rumprun' | base64)
unikernel_hypervisor=$(echo 'hvt' | base64)

umoci config --config.label="com.urunc.unikernel.binary=$unikernel_binary" --image redis-hvt:latest
umoci config --config.label="com.urunc.unikernel.cmdline=$unikernel_cmdline" --image redis-hvt:latest
umoci config --config.label="com.urunc.unikernel.unikernelType=$unikernel_type" --image redis-hvt:latest
umoci config --config.label="com.urunc.unikernel.hypervisor=$unikernel_hypervisor" --image redis-hvt:latest
```

Create a new bundle from the updated image:

```bash
umoci unpack --image redis-hvt:latest bundle2
```

Import the new bundle to ctr:

```bash
skopeo copy oci:redis-hvt:latest docker-archive:redis-hvt.tar:redis-hvt:annotated
sudo ctr images import --base-name redis-hvt:annotated redis-hvt.tar
sudo ctr image tag docker.io/library/redis-hvt:annotated harbor.nbfc.io/nubificus/redis-hvt:annotated
sudo ctr image push harbor.nbfc.io/nubificus/redis-hvt:annotated -u '$USERNAME:$PASSWORD'
```