package http

import (
	"net/http"
	"time"

	_ "github.com/ozontech/seq-ui/internal/api/httputil"
)

// serveDiffByReleases go doc.
//
//	@Router		/errorgroups/v1/diff_by_releases [post]
//	@ID			errorgroups_v1_diff_by_releases
//	@Tags		errorgroups_v1
//	@Param		body	body		diffByReleasesRequest	true	"Request body"
//	@Success	200		{object}	diffByReleasesResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error			"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetDiffByReleases(w http.ResponseWriter, r *http.Request) {
}

// nolint: unused
type diffByReleasesRequest struct {
	Service   string   `json:"service"`
	Releases  []string `json:"releases"`
	Env       *string  `json:"env,omitempty"`
	Source    *string  `json:"source,omitempty"`
	Limit     uint32   `json:"limit"`
	Offset    uint32   `json:"offset"`
	Order     order    `json:"order"`
	WithTotal bool     `json:"with_total"`
} //	@name	errorgroups.v1.DiffByReleasesRequest

// nolint: unused
type diffByReleasesResponse struct {
	Total  uint64      `json:"total"`
	Groups []diffGroup `json:"groups"`
} //	@name	errorgroups.v1.DiffByReleasesResponse

// nolint: unused
type diffGroup struct {
	Hash        string    `json:"hash" format:"uint64"`
	Message     string    `json:"message"`
	FirstSeenAt time.Time `json:"first_seen_at" format:"date-time"`
	LastSeenAt  time.Time `json:"last_seen_at" format:"date-time"`
	Source      string    `json:"source"`

	ReleaseInfos map[string]diffReleaseInfo `json:"release_infos"`
} //	@name	errorgroups.v1.DiffGroup

// nolint: unused
type diffReleaseInfo struct {
	SeenTotal uint64 `json:"seen_total"`
} //	@name	errorgroups.v1.DiffReleaseInfo
