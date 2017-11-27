#############################################
# GET/CACHE GO DEPS
#############################################
FROM golang as go-dependencies
RUN set -x \
    && go get -u github.com/revel/cmd/revel \
    && go get -u k8s.io/client-go/... \
    && go get -u k8s.io/apimachinery/... \
    && go get -u github.com/dustin/go-humanize

#############################################
# BUILD REACT APP
#############################################
FROM node:alpine as frontend
WORKDIR /app
# get npm modules (cache)
COPY ./react/package.json /app/react/
COPY ./react/package-lock.json /app/react/
RUN set -x \
    && cd /app/react \
    && npm install

# Copy app and build
COPY ./ /app
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
    && revel build k8s-devops-console /app

#############################################
# FINAL IMAGE
#############################################
FROM alpine
RUN apk add --no-cache libc6-compat
COPY --from=backend /app/ /app/
EXPOSE 9000
CMD ["/app/run.sh"]
