package types

type Role struct {
	ID          int32
	Name        string
	Permissions []string
}

type RoleInfo struct {
	Usernames []string
}

type PermissionGroup struct {
	Group       string
	Permissions []string
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
	Env       string
}

type GetRolesResponse struct {
	Roles []Role
}

type GetRoleRequest struct {
	RoleID int32
	Env    string
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
	Env      string
}

type GetUserPermissionsResponse struct {
	Permissions []string
}

type DeleteUsersFromRoleRequest struct {
	RoleID    int32
	Usernames []string
	Env       string
}
