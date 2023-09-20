TINYGO=tinygo
GOFMT=gofmt
GO=go

hello-world.wasm:  proxy_wasm_hello_world.go  
	$(GO) get
	mkdir -p ./bin
	$(TINYGO) build -o ./bin/hello-world.wasm -scheduler=none -target=wasi -tags timetzdata

clean:
	rm bin/hello-world.wasm
