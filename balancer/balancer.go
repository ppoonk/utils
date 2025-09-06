package balancer

type Balancer[T any] interface {
	Select() T
	Add(server T)
	Remove(name string)
}
