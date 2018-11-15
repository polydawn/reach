{
	"imports": {
		"base": "catalog:froob.org/base:v1:linux-amd64"
		"pipe": "catalog:example.org/proj-foo:candidate:wowslot"
	},
	"steps": {
		"main": {
			"operation": {
				"inputs": {
					"base": "/"
					"pipe": "/task/in"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"[[ $(cat /task/in) == 'hello' ]]"
					]
				}
			}
		}
	}
}
