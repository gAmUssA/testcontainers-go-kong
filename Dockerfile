ARG TC_KONG_IMAGE
FROM ${TC_KONG_IMAGE:-kong:2.8.1}

RUN mkdir -p /usr/local/kong/go-plugins/bin
