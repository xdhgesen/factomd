package interfaces

// IEtcdManager plugin interface
type IEtcdManager interface {
	SendIntoEtcd(msg []byte) error
	GetData() []byte
	Reinitiate() error
	NewBlockLease(blockHeight uint32) error
	PickUpFromHash(messageHash string) error

	// Ready will return true when the etcd client is instantiaed. It will return
	// an error if the plugin process is unreachable
	Ready() (bool, error)
}
