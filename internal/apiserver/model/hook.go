package model

import (
	"aimtp/internal/pkg/rid"
	"aimtp/pkg/authn"

	"gorm.io/gorm"
)

var (
	UserPrefix = "user"
	PostPrefix = "post"
)

// BeforeCreate 在创建数据库记录之前加密明文密码.
func (m *UserM) BeforeCreate(tx *gorm.DB) error {
	// Encrypt the user password.
	var err error
	m.Password, err = authn.Encrypt(m.Password)
	if err != nil {
		return err
	}

	return nil
}

// AfterCreate 在创建数据库记录之后生成 userID.
func (m *UserM) AfterCreate(tx *gorm.DB) error {
	m.UserID = rid.NewResourceID(UserPrefix).New(uint64(m.ID))
	return tx.Save(m).Error
}

// AfterCreate 在创建数据库记录之后生成 postID.
func (m *PostM) AfterCreate(tx *gorm.DB) error {
	m.PostID = string(rid.NewResourceID(PostPrefix).New(uint64(m.ID)))
	return tx.Save(m).Error
}
