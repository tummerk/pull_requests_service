package entity

import (
	"time"
)

const StatusOpen = "OPEN"
const StatusMerged = "MERGED"

type PullRequest struct {
	Id                string     `db:"id"`
	Name              string     `db:"name"`
	AuthorId          string     `db:"author_id"`
	Status            string     `db:"status"`
	NeedMoreReviewers bool       `db:"need_more_reviewers"`
	CreatedAt         time.Time  `db:"created_at"`
	MergedAt          *time.Time `db:"merged_at"`
	AssignedReviewers []string
}
