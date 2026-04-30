package types

type Role struct {
	ID          int32
	Name        string
	Permissions []uint64
}

type RoleRepo struct {
	ID          int32
	Name        string
	Permissions uint64
}

type Permission struct {
	Value       uint64
	Name        string
	Description *string
}

type CreateRoleRequest struct {
	Name        string
	Permissions []uint64
}

type CreateRoleRepoRequest struct {
	Name        string
	Permissions uint64
}

type CreateRoleResponse struct {
	RoleID int32
}

type AddUsersToRoleRequest struct {
	RoleID    int32
	Usernames []string
}

type GetRolesResponse struct {
	Roles                []Role
	AvailablePermissions []Permission
}

type GetRolesRepoResponse struct {
	Roles []RoleRepo
}

type GetRoleRequest struct {
	RoleID int32
}

type GetRoleResponse struct {
	Usernames []string
}

type UpdateRoleRequest struct {
	RoleID      int32
	Name        *string
	Permissions []uint64
}

type UpdateRoleRepoRequest struct {
	RoleID      int32
	Name        *string
	Permissions *uint64
}

type DeleteRoleRequest struct {
	RoleID            int32
	ReplacementRoleID *int32
}
