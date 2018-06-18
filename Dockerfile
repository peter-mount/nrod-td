# Dockerfile used to build the application

# Build container containing our pre-pulled libraries
FROM golang:alpine AS build

# The golang alpine image is missing git so ensure we have additional tools
RUN apk add --no-cache \
      curl \
      git

# We want to build our final image under /dest
# A copy of /etc/ssl is required if we want to use https datasources
#RUN mkdir -p /dest/etc &&\
#    cp -rp /etc/ssl /dest/etc/

# Ensure we have the libraries - docker will cache these between builds
RUN go get -v \
      github.com/gorilla/mux \
      github.com/peter-mount/golib/kernel \
      github.com/peter-mount/golib/rabbitmq \
      github.com/peter-mount/golib/rest \
      github.com/peter-mount/golib/statistics \
      github.com/streadway/amqp \
      gopkg.in/robfig/cron.v2 \
      gopkg.in/yaml.v2

# ============================================================
# source container contains the source as it exists within the
# repository.
FROM build AS source
WORKDIR /go/src/github.com/peter-mount/nrod-td
ADD . .

# ============================================================
FROM source AS compiler

RUN CGO_ENABLED=0 \
    GOOS=${goos} \
    GOARCH=${goarch} \
    GOARM=${goarm} \
    go build \
      -o /dest/td \
      github.com/peter-mount/nrod-td/td/bin

# Finally build the final runtime container will all required files
FROM scratch
COPY --from=compiler /dest/ /
ENTRYPOINT [ "/td" ]
CMD [ "-c", "/config.yaml" ]
