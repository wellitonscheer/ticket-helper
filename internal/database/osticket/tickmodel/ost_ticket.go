package tickmodel

import "time"

type OstTicket struct {
	TicketID    int        `db:"ticket_id"`
	TicketPID   *int       `db:"ticket_pid"`
	Number      *string    `db:"number"`
	UserID      int        `db:"user_id"`
	UserEmailID int        `db:"user_email_id"`
	StatusID    int        `db:"status_id"`
	DeptID      int        `db:"dept_id"`
	SLAID       int        `db:"sla_id"`
	TopicID     int        `db:"topic_id"`
	StaffID     int        `db:"staff_id"`
	TeamID      int        `db:"team_id"`
	EmailID     int        `db:"email_id"`
	LockID      int        `db:"lock_id"`
	Flags       int        `db:"flags"`
	Sort        int        `db:"sort"`
	IPAddress   string     `db:"ip_address"`
	Source      string     `db:"source"`
	SourceExtra *string    `db:"source_extra"`
	IsOverdue   bool       `db:"isoverdue"`
	IsAnswered  bool       `db:"isanswered"`
	DueDate     *time.Time `db:"duedate"`
	EstDueDate  *time.Time `db:"est_duedate"`
	Reopened    *time.Time `db:"reopened"`
	Closed      *time.Time `db:"closed"`
	LastUpdate  *time.Time `db:"lastupdate"`
	Created     time.Time  `db:"created"`
	Updated     time.Time  `db:"updated"`
}
