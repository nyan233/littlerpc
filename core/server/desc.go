package server

type ServiceMethod struct {
	Name    string
	Options map[string]interface{}
}

type Desc struct {
	Package        string
	ServiceMethods map[string][]ServiceMethod
}

var _ = Desc{
	Package: "memcache",
	ServiceMethods: map[string][]ServiceMethod{
		"memcache": {
			{
				Name: "Get",
				Options: map[string]interface{}{
					"NoAuth":  true,
					"gateway": true,
				},
			},
			{
				Name: "Set",
				Options: map[string]interface{}{
					"NoAuth": true,
				},
			},
			{
				Name: "Del",
				Options: map[string]interface{}{
					"NoAuth": true,
				},
			},
			{
				Name: "GetAll",
			},
		},
	},
}
