package dollop

type MsgI interface {
	Type() Type
	// Encode the frame into []byte.
	Encode() []byte

	GetData() []byte
}
