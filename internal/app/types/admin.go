package types

type Role struct {
	ID          int32
	Name        string
	Permissions []string
}

type RoleInfo struct {
	Usernames []string
}

type Permission struct {
	ID    int32
	Value string
}

type CreateRoleRequest struct {
	Name        string
	Permissions []string
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

type GetRoleRequest struct {
	RoleID int32
}

type GetRoleResponse struct {
	Usernames []string
}

type UpdateRoleRequest struct {
	RoleID      int32
	Name        *string
	Permissions []string
}

type DeleteRoleRequest struct {
	RoleID            int32
	ReplacementRoleID *int32
}

type GetUserPermissionsRequest struct {
	Username string
}

type GetUserPermissionsResponse struct {
	Permissions []string
}

type DeleteUsersFromRoleRequest struct {
	RoleID    int32
	Usernames []string
}
