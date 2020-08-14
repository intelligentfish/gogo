package restClient

type Option func(object *RestClient)

func HostOption(host string) Option {
	return func(object *RestClient) {
		object.host = host
	}
}

func PortOption(port int) Option {
	return func(object *RestClient) {
		object.port = port
	}
}
