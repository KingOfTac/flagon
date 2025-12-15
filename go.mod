module github.com/kingoftac/flagon

go 1.25.5

require (
	github.com/kingoftac/flagon/cli v0.0.0
	github.com/kingoftac/flagon/lua v0.0.0
)

require github.com/yuin/gopher-lua v1.1.1 // indirect

replace (
	github.com/kingoftac/flagon/cli => ./cli
	github.com/kingoftac/flagon/lua => ./cli-lua
)
