package model

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Account      string    `gorm:"size:64;uniqueIndex;not null" json:"account"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Realname     string    `gorm:"size:64;not null;default:''" json:"realname"`
	Email        *string   `gorm:"size:128" json:"email"`
	Status       string    `gorm:"type:enum('active','disabled');not null;default:active" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Deleted      uint8     `gorm:"not null;default:0" json:"-"`
}

func (User) TableName() string { return "user" }

type Role struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Builtin   uint8     `gorm:"not null;default:0" json:"builtin"`
	Remark    *string   `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Role) TableName() string { return "role" }

type Menu struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID  *uint64   `json:"parentId"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Type      string    `gorm:"type:enum('dir','menu','button');not null;default:menu" json:"type"`
	Path      *string   `gorm:"size:255" json:"path"`
	Icon      *string   `gorm:"size:64" json:"icon"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Status    string    `gorm:"type:enum('enabled','disabled');not null;default:enabled" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Menu) TableName() string { return "menu" }

type UserRole struct {
	UserID    uint64    `gorm:"primaryKey" json:"userId"`
	RoleID    uint64    `gorm:"primaryKey" json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (UserRole) TableName() string { return "user_role" }

type RoleMenu struct {
	RoleID    uint64    `gorm:"primaryKey" json:"roleId"`
	MenuID    uint64    `gorm:"primaryKey" json:"menuId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (RoleMenu) TableName() string { return "role_menu" }

type Product struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:110;not null" json:"name"`
	Code        *string   `gorm:"size:45;uniqueIndex" json:"code"`
	Status      string    `gorm:"type:enum('normal','closed');not null;default:normal" json:"status"`
	PO          *uint64   `gorm:"column:po" json:"po"`
	Description *string   `gorm:"type:text" json:"description"`
	CreatedBy   uint64    `gorm:"not null" json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Deleted     uint8     `gorm:"not null;default:0" json:"-"`
}

func (Product) TableName() string { return "product" }

type Story struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   uint64    `gorm:"not null;index" json:"productId"`
	Type        string    `gorm:"type:enum('planning','story');not null;default:planning" json:"type"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description *string   `gorm:"type:text" json:"description"`
	Pri         uint8     `gorm:"not null;default:3" json:"pri"`
	Status      string    `gorm:"type:enum('draft','active','closed');not null;default:draft" json:"status"`
	Estimate    *float64  `gorm:"type:decimal(10,2)" json:"estimate"`
	AssignedTo  *uint64   `json:"assignedTo"`
	OpenedBy    uint64    `gorm:"not null" json:"openedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Deleted     uint8     `gorm:"not null;default:0" json:"-"`
}

func (Story) TableName() string { return "story" }

type Project struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID   uint64     `gorm:"not null;index" json:"productId"`
	Name        string     `gorm:"size:110;not null" json:"name"`
	Code        *string    `gorm:"size:45;uniqueIndex" json:"code"`
	Status      string     `gorm:"type:enum('wait','doing','suspended','closed');not null;default:wait" json:"status"`
	Begin       *time.Time `gorm:"type:date" json:"begin"`
	End         *time.Time `gorm:"type:date" json:"end"`
	PM          *uint64    `gorm:"column:pm" json:"pm"`
	Description *string    `gorm:"type:text" json:"description"`
	CreatedBy   uint64     `gorm:"not null" json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Deleted     uint8      `gorm:"not null;default:0" json:"-"`
}

func (Project) TableName() string { return "project" }

type Sprint struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID uint64     `gorm:"not null;index" json:"projectId"`
	Name      string     `gorm:"size:110;not null" json:"name"`
	Status    string     `gorm:"type:enum('wait','doing','done','closed');not null;default:wait" json:"status"`
	Begin     *time.Time `gorm:"type:date" json:"begin"`
	End       *time.Time `gorm:"type:date" json:"end"`
	Goal      *string    `gorm:"type:text" json:"goal"`
	CreatedBy uint64     `gorm:"not null" json:"createdBy"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Deleted   uint8      `gorm:"not null;default:0" json:"-"`
}

func (Sprint) TableName() string { return "sprint" }

type SprintStory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID uint64    `gorm:"not null;index" json:"projectId"`
	SprintID  uint64    `gorm:"not null;index" json:"sprintId"`
	ProductID uint64    `gorm:"not null;index" json:"productId"`
	StoryID   uint64    `gorm:"not null;index" json:"storyId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SprintStory) TableName() string { return "sprint_story" }

type Bug struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID  uint64    `gorm:"not null;index" json:"productId"`
	ProjectID  *uint64   `json:"projectId"`
	SprintID   *uint64   `json:"sprintId"`
	StoryID    *uint64   `json:"storyId"`
	Title      string    `gorm:"size:255;not null" json:"title"`
	Steps      *string   `gorm:"type:text" json:"steps"`
	Severity   uint8     `gorm:"not null;default:3" json:"severity"`
	Pri        uint8     `gorm:"not null;default:3" json:"pri"`
	Status     string    `gorm:"type:enum('active','resolved','closed');not null;default:active" json:"status"`
	Resolution *string   `json:"resolution"`
	AssignedTo *uint64   `json:"assignedTo"`
	OpenedBy   uint64    `gorm:"not null" json:"openedBy"`
	ResolvedBy *uint64   `json:"resolvedBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Deleted    uint8     `gorm:"not null;default:0" json:"-"`
}

func (Bug) TableName() string { return "bug" }

type Attachment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ObjectType   string    `gorm:"type:enum('story','bug');not null;index:idx_attachment_object" json:"objectType"`
	ObjectID     uint64    `gorm:"not null;index:idx_attachment_object" json:"objectId"`
	OriginalName string    `gorm:"size:255;not null" json:"originalName"`
	StoredName   string    `gorm:"size:255;uniqueIndex;not null" json:"-"`
	StoragePath  string    `gorm:"size:512;not null" json:"-"`
	Ext          string    `gorm:"size:16;not null" json:"ext"`
	MimeType     *string   `gorm:"size:128" json:"mimeType"`
	SizeBytes    uint64    `gorm:"not null;default:0" json:"sizeBytes"`
	UploadedBy   uint64    `gorm:"not null" json:"uploadedBy"`
	CreatedAt    time.Time `json:"createdAt"`
	Deleted      uint8     `gorm:"not null;default:0" json:"-"`
}

func (Attachment) TableName() string { return "attachment" }
