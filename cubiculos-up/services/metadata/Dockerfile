# -------------------- STAGE 1: Build --------------------
    FROM golang:1.25 AS builder

    WORKDIR /app
    COPY ../../go.mod ../../go.sum ./
    RUN go mod download
    COPY ../../ .
    
    # Compila el ejecutable de metadata
    RUN CGO_ENABLED=0 go build -o metadata ./services/metadata/cmd
    
    # -------------------- STAGE 2: Run (Distroless) --------------------
    FROM gcr.io/distroless/base-debian12:latest
    
    WORKDIR /
    
    COPY --from=builder /app/metadata /
    
    USER nonroot:nonroot
    
    EXPOSE 50051
    
    ENTRYPOINT ["/metadata"]