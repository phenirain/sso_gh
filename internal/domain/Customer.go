package domain

import "time"


type Customer struct {
	ID               int64      `db:id`
	Email            string     `db:email"`
	FirstName        string     `db:first_name"`
	LastName         string     `db:last_name"`
	Patronymic       string     `db:patronymic"`
	Phone            string     `db:phone"`
	UserID           int64      `db:user_id"`
	CreationDatetime time.Time  `db:creation_datetime"`
	UpdateDatetime   *time.Time `db:update_datetime"`
	IsArchived       bool       `db:is_archived"`
}