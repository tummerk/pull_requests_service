package entity

import "time"

const StatusOpen = "OPEN"
const StatusMerged = "MERGED"

type PullRequest struct {
	Id                string
	Name              string
	AuthorId          string
	Status            string
	NeedMoreReviewers bool
	AssignedReviewers []string
	MergedAt          time.Time
}
