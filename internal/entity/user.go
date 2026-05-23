package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	RoleAdminID    = uuid.MustParse("019e50fb-6458-73ec-9416-2db4b3798f32")
	RoleCustomerID = uuid.MustParse("019e50fb-645a-7824-81fa-6b6e47ca998e")
)

const (
	UserVerified   = true
	UserUnverified = false
)

type Role struct {
	ID   uuid.UUID
	Name string
}

type User struct {
	Base
	RoleID      uuid.UUID
	Email       string
	Password    string
	FullName    string
	PhoneNumber string
	Address     string
	Province    string
	City        string
	PostalCode  string
	IsVerified  bool
	DeletedAt   gorm.DeletedAt

	Role Role `gorm:"foreignKey:RoleID"`
}
