#############################################
# GET/CACHE GO DEPS
#############################################
FROM golang as go-dependencies
WORKDIR /app/
COPY ./glide.yaml /app
COPY ./glide.lock /app
RUN curl https://glide.sh/get | sh && glide install
RUN go get github.com/revel/cmd/revel

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
