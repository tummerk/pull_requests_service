package value

type ExampleName string

func (e ExampleName) String() string {
	return string(e)
}
