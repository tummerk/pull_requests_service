package entity

const statusOpen = "OPEN"
const statusMerged = "MERGED"

type PullRequest struct {
	id                int
	name              string
	authorId          int
	status            string
	needMoreReviewers bool
}
