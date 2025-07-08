package entities

type role int8

const (
	Admin role = iota + 1
	Owner
	Basic
)

func (role role) ToString() string {
	var roleName string
	switch role {
	case Admin:
		roleName = "admin"
	case Basic:
		roleName = "user"
	case Owner:
		roleName = "owner"
	}
	return roleName
}

func RoleFromString(name string) role {
	var role role
	switch name {
	case "admin":
		role = Admin
	case "user":
		role = Basic
	case "owner":
		role = Owner
	}
	return role
}

type User struct {
	Id        int64
	FullName  string
	GroupName string
	GroupId   int64
	Roles     []role
	TgId      int64
}
