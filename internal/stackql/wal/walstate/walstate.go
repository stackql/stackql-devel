package walstate

type (
	WALID uint64
)

type WALState interface {
	GetWALState() string
	SetTSMState(string)
	CurrentWALID() WALID
	NextWALID() WALID
	OldestWALID() WALID
	NextUnCheckpointedWALID() (WALID, bool)
	SetCheckpointedWALID(WALID) error
}
