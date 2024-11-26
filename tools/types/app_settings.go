package types

type AppSettings interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}
