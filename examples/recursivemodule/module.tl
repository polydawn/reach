{
	"imports": {
		"base": "catalog:froob.org/base:v1:linux-amd64"
	},
	"steps": {
		"step-first": {
			"operation": {
				"inputs": {
					"base": "/"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"mkdir out; echo hello | tee /task/out/file1"
					]
				},
				"outputs": {
					"out": "/task/out"
				}
			}
		},
		"submodule-jamboree": {
			"module": {
				"imports": {
					"image": "catalog:froob.org/base:v1:linux-amd64",
					"thingy": "parent:step-first.out"
				},
				"steps": {
					"boop": {
						"operation": {
							"inputs": {
								"image": "/",
								"thingy": "/task/thingy"
							},
							"action": {
								"exec": [
									"/bin/bash", "-c",
									"mkdir out; cat thingy/file1 thingy/file1 | tee /task/out/file2"
								]
							},
							"outputs": {
								"out": "/task/out"
							}
						}
					},
					"bap": {
						"operation": {
							"inputs": {
								"image": "/"
							},
							"action": {
								"exec": [
									"/bin/bash", "-c",
									"mkdir out; echo mundus | tee /task/out/file3"
								]
							},
							"outputs": {
								"out": "/task/out"
							}
						}
					}
				},
				"exports": {
					"frob": "boop.out",
					"snoz": "bap.out"
				}
			}
		},
		"step-after": {
			"operation": {
				"inputs": {
					"base": "/",
					"submodule-jamboree.frob": "/task/frob",
					"submodule-jamboree.snoz": "/task/snoz"
				},
				"action": {
					"exec": [
						"/bin/bash", "-c",
						"mkdir out; echo cherrio, pip pip, and | cat - frob/file2 snoz/file3 | tee /task/out/product"
					]
				},
				"outputs": {
					"out": "/task/out"
				}
			}
		}
	},
	"exports": {
		"product": "step-after.out"
	}
}
