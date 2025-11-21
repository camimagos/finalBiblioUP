# ---------- STAGE 1: Build ----------
    FROM golang:1.25-bookworm AS builder

    WORKDIR /app
    
    RUN apt-get update && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*
        
    #COPY go.mod go.sum ./
    COPY ../../go.mod ../../go.sum ./

    RUN go mod download
    
    #COPY . .
    COPY ../../ .

    
    # Compile
    # Build the service
    RUN go build -o metadata ./services/metadata/cmd
    
    # ---------- STAGE 2: Run ----------
    FROM gcr.io/distroless/base-debian12:latest
    
    WORKDIR /
    
    COPY --from=builder /app/metadata .
    
    EXPOSE 50051
    
    USER nonroot:nonroot
    
    ENTRYPOINT ["/metadata"]
    