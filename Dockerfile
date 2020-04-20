FROM golang:alpine AS build

COPY bcc /src/bcc
COPY cmd /src/cmd
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum

WORKDIR /src
ENV CGO_ENABLED=0

RUN go build -o /bcc ./cmd/bcc
RUN go build -o /bcc-initdb ./cmd/bcc-initdb
RUN go build -o /bcc-github ./cmd/bcc-github

FROM scratch

COPY --from=build /bcc /bcc
COPY --from=build /bcc-initdb /bcc-initdb
COPY --from=build /bcc-github /bcc-github

ENV PATH=/
CMD ["bcc"]
