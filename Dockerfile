FROM  registry.access.redhat.com/ubi9:latest

RUN dnf install -y --nodocs skopeo && \
    dnf clean all
COPY image_resources/centos9.repo image_resources/centos9-appstream.repo /etc/yum.repos.d/
RUN dnf install -y --nodocs python3-pip python3-devel gcc && \
    dnf install -y --nodocs --nobest rsync redis && \
    dnf clean all && \
    ln -s /usr/bin/python3 /usr/bin/python

WORKDIR /workdir

RUN curl -L -o ocm-load-test-linux.tgz \
    https://github.com/cloud-bulldozer/ocm-api-load/releases/download/$(curl -L -s -H \
    'Accept: application/json' https://github.com/cloud-bulldozer/ocm-api-load/releases/latest \
    | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')/ocm-load-test-linux.tgz  && \
    tar -xf ocm-load-test-linux.tgz && \
    cp ocm-load-test /usr/local/bin/ && \
    chmod 755 /usr/local/bin/ocm-load-test

COPY config.example.yaml /workdir/config.example.yaml
RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt

CMD [ "ocm-load-test", "--config-file", "config.yaml" ]