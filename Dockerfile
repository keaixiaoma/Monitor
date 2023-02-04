FROM nvidia/cuda:10.0-base

WORKDIR /

COPY bin/manager /usr/local/bin

CMD ["manager"]
