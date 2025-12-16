module github.com/kingoftac/flagon/lua

go 1.25.5

require (
	github.com/kingoftac/flagon/cli v0.0.0
	github.com/yuin/gopher-lua v1.1.1
)

replace github.com/kingoftac/flagon/cli => ../cli
