package submission

// События модуля для outbox. Пишутся в той же транзакции, что и решение по заявке;
// письмо автору по ним формирует модуль уведомлений.
const (
	EventApproved = "submission.approved"
	EventRejected = "submission.rejected"
)

// ApprovedEvent — payload события submission.approved.
type ApprovedEvent struct {
	SubmissionID int64  `json:"submissionId"`
	AuthorID     int64  `json:"authorId"`
	Title        string `json:"title"`
	ObjectSlug   string `json:"objectSlug"`
}

// RejectedEvent — payload события submission.rejected.
type RejectedEvent struct {
	SubmissionID  int64  `json:"submissionId"`
	AuthorID      int64  `json:"authorId"`
	Title         string `json:"title"`
	Reason        string `json:"reason"`
	AuthorBlocked bool   `json:"authorBlocked"`
}
