module github.com/nickcoast/timetravel

replace github.com/temelpa/timetravel => /home/nick/code/go/temelpa

go 1.18

require github.com/gorilla/mux v1.8.0

require (
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/pelletier/go-toml v1.9.5
//github.com/temelpa/timetravel v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.5.9
	github.com/gorilla/handlers v1.5.1
)

require (
	github.com/felixge/httpsnoop v1.0.1 // indirect
	golang.org/x/exp v0.0.0-20230203172020-98cc5a0785f9
)
