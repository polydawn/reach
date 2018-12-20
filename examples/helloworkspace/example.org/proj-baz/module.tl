{
	"imports": {
		"base": "catalog:froob.org/base:v1:linux-amd64"
		"pipe": "catalog:example.org/proj-foo:v0.01:wowslot"
	},
	"steps": {
		"main": {
			"operation": {
				"inputs": {
					"/": "base",
					"/task/in": "pipe"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"[[ -f /task/in/bin/bash ]]"
					]
				}
			}
		}
	}
}
