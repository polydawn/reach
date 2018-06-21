{
	"imports": {
		"base": "catalog:froob.org/base:v1:linux-amd64"
	},
	"steps": {
		"main": {
			"operation": {
				"inputs": {
					"base": "/"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"echo hello | tee /task/out"
					]
				},
				"outputs": {
					"out": "/task/out"
				}
			}
		}
	},
	"exports": {
		"wowslot": "main.out"
	}
}
