package clog

// The clog transaction logger is a binary logger
// that logs to an abstract persistent store.
// Realistically, this is block storage.
type TxnLogger interface {
	// Log a message to the logger.
	Enqueue([]byte) (int, error)
	Commit() error
}

type LoggerManager interface {
	// Create a new transaction logger.
	NewTxnLogger() (TxnLogger, error)
	// Close the logger manager.
	Close() error
}
