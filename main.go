package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	UID          int64
	Username     string
	Email        string
	Phone        *string
	PasswordHash string
	Status       int8
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt
}

// ---------------------------------------------------------------------------
// DAO (Data Access Object)
// ---------------------------------------------------------------------------

// UserDAO 封装用户表 CRUD.
// 简单的分层: 所有方法接收 *gorm.DB, 便于在事务或不同 shard 上复用.
type UserDAO struct{}

// Create 插入单条记录 (OnConflict 忽略冲突主键).
func (UserDAO) Create(db *gorm.DB, user *User) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(user).Error
}

// BatchCreate 批量插入.
func (UserDAO) BatchCreate(db *gorm.DB, users []User) error {
	return db.CreateInBatches(users, 100).Error
}

// GetByID 按主键查询 (自动处理软删除).
func (UserDAO) GetByID(db *gorm.DB, id int64) (*User, error) {
	var u User
	err := db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername 按用户名查询.
func (UserDAO) GetByUsername(db *gorm.DB, username string) (*User, error) {
	var u User
	err := db.Where("username = ?", username).First(&u).Error
	return &u, err
}

// List 分页+条件查询, 返回用户列表和总数.
//
//	filters: WHERE 条件, 如 map[string]any{"status": 1}
//	preload: 若有关联表可预加载 (后续多表查询用)
func (UserDAO) List(db *gorm.DB, filters map[string]any, offset, limit int) ([]User, int64, error) {
	var users []User
	var total int64

	tx := db.Model(&User{})
	for k, v := range filters {
		tx = tx.Where(k, v)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Offset(offset).Limit(limit).Order("id ASC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// Update 按主键更新指定字段.
func (UserDAO) Update(db *gorm.DB, id int64, values map[string]any) error {
	return db.Model(&User{}).Where("id = ?", id).Updates(values).Error
}

// Delete 软删除 (GORM 默认软删除).
func (UserDAO) Delete(db *gorm.DB, id int64) error {
	return db.Delete(&User{}, id).Error
}

// HardDelete 物理删除 (分片合并/数据清理时用).
func (UserDAO) HardDelete(db *gorm.DB, id int64) error {
	return db.Unscoped().Delete(&User{}, id).Error
}

func main() {
	dsn := "dbname=postgres sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	dao := UserDAO{}

	// ---- Create ----
	user := User{
		UID:          1700000000000000, // snowflake ID
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "$2a$10$...",
		Status:       1,
	}
	if err := dao.Create(db, &user); err != nil {
		log.Printf("Create: %v", err)
	}

	// ---- BatchCreate ----
	batch := make([]User, 0, 3)
	for i := int64(1); i <= 3; i++ {
		batch = append(batch, User{
			UID:          1700000000000000 + i,
			Username:     fmt.Sprintf("user_%d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "$2a$10$...",
			Status:       1,
		})
	}
	if err := dao.BatchCreate(db, batch); err != nil {
		log.Printf("BatchCreate: %v", err)
	}

	// ---- GetByID ----
	u, err := dao.GetByID(db, 1700000000000001)
	if err != nil {
		log.Printf("GetByID: %v", err)
	} else {
		log.Printf("GetByID: %+v", *u)
	}

	// ---- GetByUsername ----
	u, err = dao.GetByUsername(db, "alice")
	if err != nil {
		log.Printf("GetByUsername: %v", err)
	} else {
		log.Printf("GetByUsername: %+v", *u)
	}

	// ---- List ----
	users, total, err := dao.List(db, map[string]any{"status": 1}, 0, 10)
	if err != nil {
		log.Printf("List: %v", err)
	} else {
		log.Printf("List: total=%d, users=%+v", total, users)
	}

	// ---- Update ----
	if err := dao.Update(db, 1700000000000001, map[string]any{
		"nickname": "Alice Updated",
		"avatar":   "https://example.com/avatar.png",
	}); err != nil {
		log.Printf("Update: %v", err)
	}

	// ---- Delete (soft) ----
	if err := dao.Delete(db, 1700000000000004); err != nil {
		log.Printf("Delete: %v", err)
	}

	// 验证软删除: 再查已删除的记录应返回 ErrRecordNotFound
	if _, err := dao.GetByID(db, 1700000000000004); err != nil {
		log.Printf("Soft-deleted user not found (expected): %v", err)
	}

	// Unscoped 可以查到被软删除的记录
	var deletedUser User
	if err := db.Unscoped().First(&deletedUser, 1700000000000004).Error; err == nil {
		log.Printf("Unscoped found deleted user: %+v", deletedUser)
	}
}
