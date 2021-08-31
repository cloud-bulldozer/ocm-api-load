FROM python:3.9.6-bullseye

VOLUME [ "/results" ]

COPY ./build/ocm-load-test automation.py requirements.txt /workdir/
COPY config.example.yaml /workdir/config.yaml

WORKDIR /workdir
RUN pip3 install -r requirements.txt

CMD [ "./ocm-load-test", "--config-file", "config.yaml" ]