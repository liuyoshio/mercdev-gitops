# ---- build stage ----
FROM golang:1.26.3 AS build
WORKDIR /src
# cache deps first: only re-runs when go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO disabled => fully static binary, runs on a scratch/distroless base
RUN CGO_ENABLED=0 GOOS=linux go build -o /platformd .

# ---- final stage ----
FROM gcr.io/distroless/static-debian12
COPY --from=build /platformd /platformd
EXPOSE 50051 8080
USER nonroot:nonroot
ENTRYPOINT ["/platformd"]
CMD ["serve"]