package tsm

type TransactionStorageManager interface {
	// Orchestrate:
	//   1. Lock Manager.
	//   2. Access Methods.
	//   3. Log Manager.
	//   4. Buffer Manager.
}

type LockManager interface {
	// - Locking through runtime libs.
	// - Latches on
}

type AccessMethods interface {
	// Probably heterogeneous, so this will be some sort of monoliuthic
	// union of various access method intercases
	// - HTTP access
	// - RDBMS access, SQL queries in general
	// - Auth
	// - [FUTURE] Access to statically linked `C` and `golang` libs, eg stdlib stuff, os libs...
	// - [FUTURE] Access to dynamically linked libraries, do NOT favour support for the hashicorp `golang` plugin system.
	// - [FUTURE] Access to arbitrary TCP/UDP services.
}

type LogManager interface {
	// Must contain sufficient information to:
	//   - Authenticate.
	//   - Run access method calls in the correct order.
	//   - Support the appropriate isolation level.
}

type BufferManager interface {
	// Might well be N/A for now.
}

type HTTPAccess interface {
	// A portion of the AccessMethods interface.
	// Pretty much a read only version of what is currently
	// defined in:
	//   - `go-openapistackql.HTTPArmoury`.
	//   - `go-openapistackql.HTTPArmouryParameters`.
	//   - `go-openapistackql.HttpParameters`.
}
