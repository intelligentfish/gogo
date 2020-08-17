package restClient

// Option option
type Option func(object *RestClient)

// HostOption host option
func HostOption(host string) Option {
	return func(object *RestClient) {
		object.host = host
	}
}

// Port port option
func PortOption(port int) Option {
	return func(object *RestClient) {
		object.port = port
	}
}
