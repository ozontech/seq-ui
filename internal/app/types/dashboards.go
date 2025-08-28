package types

type DashboardInfo struct {
	UUID string
	Name string
}

type DashboardInfos []DashboardInfo

type DashboardInfoWithOwner struct {
	DashboardInfo
	OwnerName string
}

type DashboardInfosWithOwner []DashboardInfoWithOwner

type Dashboard struct {
	Name      string
	Meta      string
	OwnerName string
}

// temporarily to maintain compatibility
type GetDashboardsRequest struct {
	ProfileID int64
}

type GetAllDashboardsRequest struct {
	Limit  int
	Offset int
}

type GetUserDashboardsRequest struct {
	ProfileID int64
	Limit     int
	Offset    int
}

type CreateDashboardRequest struct {
	ProfileID int64
	Name      string
	Meta      string
}

type UpdateDashboardRequest struct {
	UUID      string
	ProfileID int64
	Name      *string
	Meta      *string
}

func (ur UpdateDashboardRequest) IsEmpty() bool {
	return ur.Name == nil && ur.Meta == nil
}

type DeleteDashboardRequest struct {
	UUID      string
	ProfileID int64
}

type SearchDashboardsFilter struct {
	OwnerName *string
}

type SearchDashboardsRequest struct {
	Query  string
	Limit  int
	Offset int
	Filter *SearchDashboardsFilter
}
