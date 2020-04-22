package cfgtool

type Status struct {
	Rings []Ring
	NodeId string
}

type Ring struct {
	Id      string
	Address string
	Faulty  bool
}
