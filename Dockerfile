#############################################
# GET/CACHE GO DEPS
#############################################
FROM golang as go-dependencies
RUN go get -u github.com/revel/cmd/revel
RUN go get -u k8s.io/client-go/...
RUN go get -u k8s.io/apimachinery/...
RUN go get -u golang.org/x/oauth2
RUN go get -u github.com/dustin/go-humanize
RUN go get -u cloud.google.com/go/compute/metadata
RUN go get -u github.com/google/go-github/github
RUN go get -u github.com/coreos/go-oidc
RUN go get -u gopkg.in/yaml.v2

#############################################
# GET/CACHE NPM DEPS
#############################################
FROM node:alpine as npm-dependencies
WORKDIR /app
# get npm modules (cache)
COPY ./react/package.json /app/react/
COPY ./react/package-lock.json /app/react/
RUN set -x \
    && cd /app/react \
    && npm install

#############################################
# BUILD REACT APP
#############################################
FROM node:alpine as frontend
# Copy app and build
COPY ./ /app
COPY --from=npm-dependencies /app/react/node_modules/ /app/react/node_modules/
RUN set -x \
    && cd /app/react/ \
    && npm run build \
    && cp /app/react/build/index.html /app/app/views/react.html \
    && rm -rf /app/public/js \
    && mkdir -p /app/public/js \
    && cp /app/react/build/static/js/* /app/public/js \
    && rm -rf /app/react

#############################################
# BUILD GO APP
#############################################
FROM golang as backend
COPY --from=go-dependencies /go /go
COPY --from=frontend /app /go/src/k8s-devops-console
RUN set -x \
    && revel build k8s-devops-console /app prod \
    && cp /go/src/k8s-devops-console/docker/entrypoint.sh /app/entrypoint.sh \
    && chmod +x /app/entrypoint.sh \
    && rm -rf /app/src/k8s-devops-console/tests

#############################################
# FINAL IMAGE
#############################################
FROM alpine
RUN apk add --no-cache \
        libc6-compat \
    	ca-certificates
COPY --from=backend /app/ /app/
EXPOSE 9000
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/run.sh"]
