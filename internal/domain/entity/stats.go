package entity

type UserAssignmentStat struct {
	UserID          string `db:"user_id"`
	Username        string `db:"username"`
	AssignmentCount int    `db:"assignment_count"`
}
