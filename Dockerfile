FROM docker.io/alpine:3.14

ARG TARGETARCH=amd64

RUN apk add --no-cache \
      python3 py3-pip py3-setuptools py3-wheel \
      py3-pillow \
      py3-aiohttp \
      py3-magic \
      py3-sqlalchemy \
      py3-psycopg2 \
      py3-ruamel.yaml \
      py3-commonmark \
      py3-alembic \
      py3-prometheus-client \
      #hangups
        py3-async-timeout \
        py3-requests \
        py3-appdirs \
        py3-configargparse \
        py3-protobuf \
        py3-urwid \
        py3-lxml \
        #mechanicalsoup
          py3-beautifulsoup4 \
      py3-idna \
      # encryption
      py3-olm \
      py3-cffi \
      py3-pycryptodome \
      py3-unpaddedbase64 \
      py3-future \
      # proxy support
      py3-pysocks \
      # Other dependencies
      ca-certificates \
      su-exec \
      bash \
      curl \
      jq \
      yq \
      git

COPY requirements.txt /opt/mautrix-googlechat/requirements.txt
COPY optional-requirements.txt /opt/mautrix-googlechat/optional-requirements.txt
WORKDIR /opt/mautrix-googlechat
RUN apk add --virtual .build-deps python3-dev libffi-dev build-base \
 && sed -Ei 's/psycopg2-binary.+//' optional-requirements.txt \
 && pip3 install -r requirements.txt -r optional-requirements.txt \
 && apk del .build-deps

COPY . /opt/mautrix-googlechat
RUN pip3 install .[e2be] \
  # This doesn't make the image smaller, but it's needed so that the `version` command works properly
  && cp mautrix_googlechat/example-config.yaml . && rm -rf mautrix_googlechat

ENV UID=1337 GID=1337
VOLUME /data

CMD ["/opt/mautrix-googlechat/docker-run.sh"]
