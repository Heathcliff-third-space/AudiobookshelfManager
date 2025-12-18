package models

// ServerStatus 服务器状态信息
type ServerStatus struct {
	Success bool   `json:"success"`
	ServerVersion string `json:"serverVersion"`
	APIVersion string `json:"apiVersion"`
	UserID string `json:"userId,omitempty"`
	Username string `json:"username,omitempty"`
	Language string `json:"language,omitempty"`
}

// LibraryInfo 媒体库信息
type LibraryInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Folders     []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	} `json:"folders"`
	DisplayOrder int64       `json:"displayOrder"`
	Icon         string      `json:"icon"`
	LastScan     int64       `json:"lastScan,omitempty"` // 修改为int64
	CreatedAt    int64       `json:"createdAt"`
	UpdatedAt    int64       `json:"updatedAt"`
	MediaType    string      `json:"mediaType"`
	Provider     string      `json:"provider"`
	Settings     interface{} `json:"settings"`
}

// Book 书籍信息
// 根据API响应，我们只需要用到path、size、addedAt字段
// 这些字段来自libraryItem对象
// 现在添加libraryId字段以显示对应的媒体库，并使用relPath代替path以提高安全性
type Book struct {
	LibraryID string `json:"libraryId"`
	RelPath   string `json:"relPath"`
	Size      int64  `json:"size"`
	AddedAt   int64  `json:"addedAt"`
}

// ServerInfo 服务器基本信息
type ServerInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	PublicIP      string `json:"publicIP"`
	LocalIP       string `json:"localIP"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	StartTime     int64  `json:"startTime"`
	Uptime        int64  `json:"uptime"`
	TotalRAM      int64  `json:"totalRAM"`
	FreeRAM       int64  `json:"freeRAM"`
	TotalDiskSize int64  `json:"totalDiskSize"`
	FreeDiskSize  int64  `json:"freeDiskSize"`
}

// AuthFormData 认证表单数据
type AuthFormData struct {
	AuthLoginCustomMessage interface{} `json:"authLoginCustomMessage"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Type          string `json:"type"`
	Token         string `json:"token,omitempty"`
	IsActive      bool   `json:"isActive"`
	LastSeen      int64  `json:"lastSeen"`
	Permissions   Permissions `json:"permissions"`
	MediaProgress []interface{} `json:"mediaProgress"` // 根据实际情况调整类型
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// Permissions 用户权限
type Permissions struct {
	Download bool `json:"download"`
	Update   bool `json:"update"`
	Delete   bool `json:"delete"`
	Upload   bool `json:"upload"`
	AccessAllLibraries bool `json:"accessAllLibraries"`
	AccessAllTags      bool `json:"accessAllTags"`
	AccessExplicitContent bool `json:"accessExplicitContent"`
}

// Settings 服务器设置
type Settings struct {
	ID                    string `json:"id"`
	ScannerFindCovers     bool   `json:"scannerFindCovers"`
	ScannerCoverProvider  string `json:"scannerCoverProvider"`
	ScannerParseSubtitle  bool   `json:"scannerParseSubtitle"`
	ScannerPreferMatchedMetadata bool `json:"scannerPreferMatchedMetadata"`
	ScannerDisableWatcher bool   `json:"scannerDisableWatcher"`
}