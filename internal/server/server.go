package server

type Server struct {
	ExampleServer
}

func NewServer(
	exampleServer ExampleServer,
) Server {
	return Server{
		ExampleServer: exampleServer,
	}
}
