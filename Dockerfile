ARG TC_KONG_IMAGE
FROM ${TC_KONG_IMAGE:-kong/kong:3.4.0}

RUN mkdir -p /usr/local/kong/go-plugins/bin
