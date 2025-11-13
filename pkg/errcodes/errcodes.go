package errcodes

import "git.appkode.ru/pub/go/failure"

const (
	TeamAlreadyExists   failure.ErrorCode = "TeamAlreadyExists"
	UserAlreadyExists   failure.ErrorCode = "UserAlreadyExists"
	PullRequestExists   failure.ErrorCode = "PullRequestExists"
	NotFound            failure.ErrorCode = "NotFound"
	InternalServerError failure.ErrorCode = "InternalError"
)
