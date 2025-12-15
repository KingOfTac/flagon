$env:GOOS = 'js'
$env:GOARCH = 'wasm'
go build -o ./examples/public/flagon.wasm ./wasm