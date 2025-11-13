package errcodes

import "git.appkode.ru/pub/go/failure"

const (
	TeamAlreadyExists   failure.ErrorCode = "TEAM_EXISTS"
	UserAlreadyExists   failure.ErrorCode = "USER_EXISTS"
	PullRequestExists   failure.ErrorCode = "PR_EXISTS"
	NotFound            failure.ErrorCode = "NOT_FOUND"
	InternalServerError failure.ErrorCode = "INTERNAL_SERVER_ERROR"
	NoCandidate         failure.ErrorCode = "NO_CANDIDATE"
	PrMerged            failure.ErrorCode = "PR_MERGED"
	NotAssigned         failure.ErrorCode = "NOT_ASSIGNED"
)
