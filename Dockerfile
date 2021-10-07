FROM registry.access.redhat.com/ubi8:latest

COPY image_resources/centos8-appstream.repo /etc/yum.repos.d/centos8-appstream.repo
RUN dnf install -y --nodocs python3 python3-pip && dnf clean all
RUN dnf install -y --nodocs skopeo redis curl --enablerepo=centos8-appstream && dnf clean all
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN useradd -mUs /bin/bash api
WORKDIR /home/api/workdir

RUN curl -L -o ocm-load-test-linux.tgz \
    https://github.com/cloud-bulldozer/ocm-api-load/releases/download/$(curl -L -s -H \
    'Accept: application/json' https://github.com/cloud-bulldozer/ocm-api-load/releases/latest \
    | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')/ocm-load-test-linux.tgz; \
    tar -xf ocm-load-test-linux.tgz

RUN chown -R api:api /home/api/workdir

RUN cp ocm-load-test /usr/local/bin/
RUN chmod 755 /usr/local/bin/ocm-load-test

COPY config.example.yaml /home/api/workdir/config.yaml
USER api
ENV PATH="/home/api/.local/bin:${PATH}"
RUN pip3 install --user -r requirements.txt

CMD [ "ocm-load-test", "--config-file", "config.yaml" ]