package protocol

type Serverbound interface {
	ID() int32
	Decode(*Reader)
}
type Clientbound interface {
	ID() int32
	Encode(*Writer)
}
