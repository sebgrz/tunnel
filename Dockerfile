FROM golang:1.18
WORKDIR build
COPY . /build
RUN cd /build && find -L cmd/* -type d -exec go build -tags musl -o /app/{} {}/main.go \;
CMD /app/cmd/server
