package entity

const statusOpen = "OPEN"
const statusMerged = "MERGED"

type PullRequest struct {
	Id                string
	Name              string
	AuthorId          string
	Status            string
	NeedMoreReviewers bool
	AssignedReviewers []string
}
