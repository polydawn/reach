{
	"imports": {
		"base": "catalog:images.polydawn.io/busybash:v1:linux-amd64",
		"go": "catalog:entwine.polydawn.io/go:v1.10:linux-amd64",
		"src": "ingest:git:.:HEAD"
	},
	"steps": {
		"build": {
			"operation": {
				"inputs": {
					"base": "/",
					"go": "/app/go",
					"src": "/task"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"export PATH=$PATH:/app/go/go/bin && export GOPATH=$PWD/.gopath && export GOBIN=$PWD/out/bin && go install ./cmd/*"
					]
				},
				"outputs": {
					"bin": "/task/out"
				}
			}
		},
		"test": {
			"operation": {
				"inputs": {
					"base": "/",
					"build.bin": "/task"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"./bin/stellar -h"
					]
				}
			}
		}
	},
	"exports": {
		"bin-linux-amd64": "build.bin",
		"src": "src"
	}
}
