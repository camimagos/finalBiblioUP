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

    
    # Build the service
    RUN go build -o reservation ./services/reservation/cmd
    
    # ---------- STAGE 2: Run ----------
    FROM gcr.io/distroless/base-debian12:latest
    
    WORKDIR /
    
    COPY --from=builder /app/reservation .
    
    EXPOSE 50052
    
    USER nonroot:nonroot
    
    ENTRYPOINT ["/reservation"]
    