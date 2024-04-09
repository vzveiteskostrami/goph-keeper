package co

import "time"

const (
	SessionNotDefined = iota + 1
	SessionLocal
	SessionBoth
	SessionClose
)

const (
	EntityNotDefined int16 = iota
	EntityLoginPassword
	EntityCard
	EntityText
	EntityBinary
)

const (
	StrictNo int16 = iota
	StrictRead
	StrictWrite
)

const ServerKey string = "L_vWeAMKmC1yTOThJtwA9g1$LN1zqTou-$*o8J$G596N459ERBNV_7340t3_47b4"

type Udata struct {
	Oid        *int64     `json:"oid,omitempty"`
	UserId     *int64     `json:"user_id,omitempty"`
	DataType   *int16     `json:"data_type,omitempty"`
	DataName   *string    `json:"data_name,omitempty"`
	CreateTime *time.Time `json:"create_time,omitempty"`
	UpdateTime *time.Time `json:"update_time,omitempty"`
	DeleteFlag *bool      `json:"delete_flag,omitempty"`
	Data       *string    `json:"data,omitempty"`
}

type RequestList struct {
	Full *bool   `json:"full,omitempty"`
	All  *bool   `json:"all,omitempty"`
	Data []Udata `json:"data,omitempty"`
}
