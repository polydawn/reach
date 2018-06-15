{
	"imports": {
		"base": "catalog:base:v1:linux-amd64"
	},
	"steps": {
		"main": {
			"operation": {
				"inputs": {
					"base": "/"
				},
				"action": {
					"exec": [
						"/bin/bash",
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
