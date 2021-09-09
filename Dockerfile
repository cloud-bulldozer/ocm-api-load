FROM python:3.9.6-bullseye

USER root
RUN useradd -mUs /bin/bash api

COPY ./build/ocm-load-test /usr/local/bin/
RUN chmod 755 /usr/local/bin/ocm-load-test

COPY automation.py requirements.txt /home/api/workdir/
COPY config.example.yaml /home/api/workdir/config.yaml
WORKDIR /home/api/workdir
RUN chown -R api:api /home/api/workdir
USER api
ENV PATH="/home/api/.local/bin:${PATH}"
RUN pip3 install -r requirements.txt

CMD [ "ocm-load-test", "--config-file", "config.yaml" ]