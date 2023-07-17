package server

type OracleServer interface {
	Get([]byte) []byte
}
