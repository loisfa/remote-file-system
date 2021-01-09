module github.com/loisfa/remote-file-system/api

go 1.15

// can be needed on windows if do path not properly defined
// replace github.com/loisfa/remote-file-system/api => c:/Users/{user}/go/src/github.com/loisfa/remote-file-system/api

require (
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
)
