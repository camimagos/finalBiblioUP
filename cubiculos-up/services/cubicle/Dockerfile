# -------------------- STAGE 1: Build --------------------
    FROM golang:1.25 AS builder

    # Establece el directorio de trabajo donde estarán los fuentes
    WORKDIR /app
    
    # Copia los archivos de gestión de módulos (go.mod, go.sum)
    # Es mejor copiarlos individualmente y no el directorio padre
    COPY ../../go.mod ../../go.sum ./
    
    # Descarga las dependencias
    RUN go mod download
    
    # Copia el resto del proyecto (incluyendo 'proto' y 'services')
    COPY ../../ .
    
    # Compila el ejecutable del servicio. Usamos un 'go build' más explícito.
    # CGO_ENABLED=0 asegura que el binario sea estático, ideal para Distroless.
    RUN CGO_ENABLED=0 go build -o cubicle ./services/cubicle/cmd
    
    # -------------------- STAGE 2: Run (Distroless) --------------------
    FROM gcr.io/distroless/static-debian12:latest
    
    # Usa 'static' en lugar de 'base' si CGO está deshabilitado. Es aún más pequeño.
    # Nota: Si tu código usa librerías C (como el driver de Postgres), debes volver a 'base'. 
    # Ya que Metadata y Reservation se conectan a Postgres, usaremos 'base' para la plantilla general.
    
    FROM gcr.io/distroless/base-debian12:latest
    
    WORKDIR /
    
    # Copia el binario compilado de la etapa 'builder'
    COPY --from=builder /app/cubicle /
    
    # Establece el usuario no root para seguridad
    USER nonroot:nonroot
    
    # Expone el puerto (documentación)
    EXPOSE 50053
    
    # Comando de ejecución
    ENTRYPOINT ["/cubicle"]