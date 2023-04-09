package acid_dto //nolint:revive,stylecheck // meaning is clear

import (
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
)

var (
	_ CommitCoDomain = &standardCommitCoDomain{}
)

type CommitCoDomain interface {
	GetExecutorOutput() []internaldto.ExecutorOutput
	// getVotingPhaseError() error
	// getCompletionPhaseError() error
	GetError() error
	GetMessages() []string
	IsErroneous() bool
}

type standardCommitCoDomain struct {
	executorOutputs      []internaldto.ExecutorOutput
	votingPhaseError     error
	completionPhaseError error
}

func NewCommitCoDomain(
	executorOutputs []internaldto.ExecutorOutput,
	votingPhaseError error,
	completionPhaseError error,
) CommitCoDomain {
	return &standardCommitCoDomain{
		executorOutputs:      executorOutputs,
		votingPhaseError:     votingPhaseError,
		completionPhaseError: completionPhaseError,
	}
}

func (c *standardCommitCoDomain) GetMessages() []string {
	var messages []string
	if c.votingPhaseError != nil {
		messages = append(messages, c.votingPhaseError.Error())
	}
	if c.completionPhaseError != nil {
		messages = append(messages, c.completionPhaseError.Error())
	}
	return messages
}

func (c *standardCommitCoDomain) IsErroneous() bool {
	return c.votingPhaseError != nil || c.completionPhaseError != nil
}

func (c *standardCommitCoDomain) GetExecutorOutput() []internaldto.ExecutorOutput {
	return c.executorOutputs
}

func (c *standardCommitCoDomain) GetError() error {
	if c.votingPhaseError != nil {
		return c.votingPhaseError
	}
	return c.completionPhaseError
}

// func (c *standardCommitCoDomain) getVotingPhaseError() error {
// 	return c.votingPhaseError
// }

// func (c *standardCommitCoDomain) getCompletionPhaseError() error {
// 	return c.completionPhaseError
// }
