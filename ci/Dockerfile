FROM python:3.10.12-slim-bullseye

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y

RUN DEBIAN_FRONTEND=noninteractive apt-get install curl -y

# Install systemctl.py
RUN curl -L -o systemctl https://raw.githubusercontent.com/gdraheim/docker-systemctl-replacement/master/files/docker/systemctl3.py
RUN mv systemctl /usr/bin/systemctl
RUN chmod +x /usr/bin/systemctl
RUN ln -s /usr/local/bin/python3  /usr/bin/python3

WORKDIR /temporary

# Install containerd binaries
RUN CTR_V=$(curl -L -s -o /dev/null -w '%{url_effective}' "https://github.com/containerd/containerd/releases/latest" | \
        grep -oP "v\d+\.\d+\.\d+" | sed 's/v//') &&\
    curl -L -o containerd.tar.gz \
        https://github.com/containerd/containerd/releases/latest/download/containerd-$CTR_V-linux-$(dpkg --print-architecture).tar.gz

RUN tar Cxzvf /usr/local containerd.tar.gz && \
    rm -f containerd.tar.gz

# Install containerd service
RUN mkdir -p /lib/systemd/system && \
    CTR_V=$(curl -L -s -o /dev/null -w '%{url_effective}' "https://github.com/containerd/containerd/releases/latest" | \
        grep -oP "v\d+\.\d+\.\d+" | sed 's/v//') &&\
    curl -L -o containerd.service https://raw.githubusercontent.com/containerd/containerd/v$CTR_V/containerd.service && \
    mv containerd.service /lib/systemd/system/containerd.service

RUN sed -i 's/^ExecStartPre=-\/sbin\/modprobe overlay/#&/' /lib/systemd/system/containerd.service


# Enable containerd service
RUN systemctl daemon-reload && systemctl enable containerd

# Cleanup
WORKDIR /
RUN rm -fr /temporary

CMD ["bin/bash"]
# ENTRYPOINT [ "/usr/bin/systemctl" ]