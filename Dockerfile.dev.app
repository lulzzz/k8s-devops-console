#############################################
# GET/CACHE GO DEPS
#############################################
FROM golang
RUN go get -u github.com/revel/cmd/revel
RUN go get -u k8s.io/client-go/...
RUN go get -u k8s.io/apimachinery/...
RUN go get -u golang.org/x/oauth2
RUN go get -u github.com/dustin/go-humanize
RUN go get -u cloud.google.com/go/compute/metadata
RUN go get -u github.com/google/go-github/github
RUN go get -u github.com/coreos/go-oidc
EXPOSE 9000
CMD ["revel", "run", "k8s-devops-console"]
