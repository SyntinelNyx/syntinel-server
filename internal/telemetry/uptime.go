package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/jackc/pgx/v5/pgtype"
)

type UptimeResponse struct {
	CheckTime        time.Time
	TotalAssets      int64
	AssetsUp         interface{}
	UptimePercentage int32
}

func (h *Handler) Uptime(w http.ResponseWriter, r *http.Request) {
	var rootId pgtype.UUID
	var err error

	account := auth.GetClaims(r.Context())
	if account.AccountType != "root" {
		rootId, err = h.queries.GetRootAccountIDForIAMUser(context.Background(), account.AccountID)
		if err != nil {
			response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to get associated root account for IAM account", err)
			return
		}
	} else {
		rootId = account.AccountID
	}

	uptimeData, err := h.queries.GetAssetsUpByHour(context.Background(), rootId)
	if err != nil {
		response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve asset uptime data", err)
		return
	}

	parsedData := parseUptimeData(uptimeData)

	// Respond with the uptime data
	response.RespondWithJSON(w, http.StatusOK, parsedData)
}

func parseUptimeData(data []query.GetAssetsUpByHourRow) []UptimeResponse {
	var parsedData []UptimeResponse

	for _, entry := range data {
		parsedEntry := UptimeResponse{
			CheckTime:        time.Unix(entry.CheckTime, 0),
			TotalAssets:      entry.TotalAssets,
			AssetsUp:         entry.AssetsUp,
			UptimePercentage: entry.UptimePercentage,
		}
		parsedData = append(parsedData, parsedEntry)
	}

	return parsedData
}


	// -- name: GetAssetsUpByHour :many
	// WITH hour_buckets AS (
	//   -- Generate series of hourly buckets for the last 30 days
	//   SELECT generate_series(
	// 	date_trunc('hour', NOW() - INTERVAL '30 days'),
	// 	date_trunc('hour', NOW()),
	// 	INTERVAL '1 hour'
	//   ) AS bucket_time
	// ),
	// asset_listing AS (
	//   -- Get distinct list of assets for the root account
	//   SELECT DISTINCT asset_id
	//   FROM telemetry_asset
	//   WHERE telemetry_asset.root_account_id = 'a2fce64d-4620-43c5-a988-4fd2ce7984b1'
	// ),
	// telemetry_with_lag AS (
	//   -- Calculate lag times in a separate CTE
	//   SELECT
	// 	ta.asset_id,
	// 	t.telemetry_time,
	// 	LAG(t.telemetry_time) OVER (PARTITION BY ta.asset_id ORDER BY t.telemetry_time) AS prev_telemetry_time,
	// 	EXTRACT(EPOCH FROM (t.telemetry_time - LAG(t.telemetry_time) OVER (PARTITION BY ta.asset_id ORDER BY t.telemetry_time))) AS time_diff
	//   FROM
	// 	telemetry_asset ta
	//   JOIN
	// 	telemetry t ON ta.telemetry_id = t.telemetry_id
	//   WHERE
	// 	ta.root_account_id = 'a2fce64d-4620-43c5-a988-4fd2ce7984b1'
	// 	AND t.telemetry_time > NOW() - INTERVAL '30 days'
	// ),
	// asset_uptime_status AS (
	//   -- Calculate uptime status for each asset in each hour
	//   SELECT
	// 	date_trunc('hour', telemetry_time) AS hour,
	// 	asset_id,
	// 	1 AS asset_counted,
	// 	CASE
	// 	  WHEN MAX(CASE WHEN time_diff <= 300 OR prev_telemetry_time IS NULL THEN 1 ELSE 0 END) = 1 THEN 1
	// 	  ELSE 0
	// 	END AS was_up
	//   FROM
	// 	telemetry_with_lag
	//   GROUP BY
	// 	date_trunc('hour', telemetry_time), asset_id
	// ),
	// hour_asset_matrix AS (
	//   -- Create a complete matrix of all hours and all assets
	//   SELECT
	// 	h.bucket_time,
	// 	a.asset_id
	//   FROM
	// 	hour_buckets h
	//   CROSS JOIN
	// 	asset_listing a
	// )
	// SELECT
	//   hm.bucket_time AS check_time,
	//   COUNT(hm.asset_id) AS total_assets,
	//   COALESCE(SUM(aus.was_up), 0) AS assets_up,
	//   CASE
	// 	WHEN COUNT(hm.asset_id) > 0
	// 	THEN ROUND((COALESCE(SUM(aus.was_up), 0)::numeric / COUNT(hm.asset_id)) * 100, 2)
	// 	ELSE 0
	//   END AS uptime_percentage
	// FROM
	//   hour_asset_matrix hm
	// LEFT JOIN
	//   asset_uptime_status aus ON hm.bucket_time = aus.hour AND hm.asset_id = aus.asset_id
	// GROUP BY
	//   hm.bucket_time
	// ORDER BY
	//   hm.bucket_time DESC;
