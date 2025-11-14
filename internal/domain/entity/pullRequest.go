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
	CreatedAt         time.Time
	MergedAt          time.Time
}
