FROM ubuntu AS build-stage

RUN apt update && apt upgrade -y && apt install -y curl

WORKDIR /usr

RUN curl https://wazero.io/install.sh | sh

FROM alpine

COPY --from=build-stage /usr/bin/wazero /usr/bin/wazero

RUN apk update && apk add --no-cache gzip

COPY wasmbox /usr/bin/wasmbox

ENTRYPOINT ["sh", "/usr/bin/wasmbox"]


