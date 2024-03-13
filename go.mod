module github.com/rgtexa/boot-gohttpserver

go 1.22.0

require github.com/go-chi/chi/v5 v5.0.12

require internal/godb v1.0.0

require golang.org/x/crypto v0.21.0 // indirect

replace internal/godb => ./internal/godb
