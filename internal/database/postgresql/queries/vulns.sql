-- name: InsertNewVulnerabilities :exec
INSERT INTO vulnerability_data(vulnerability_id)
SELECT unnest(@vuln_list::text []) ON CONFLICT (vulnerability_id) DO NOTHING;

-- name: RetrieveUnchangedVulnerabilities :many
SELECT vd.vulnerability_id
FROM unnest(@vuln_list::text []) WITH ORDINALITY AS vuln(elem, ord)
    JOIN unnest(@modified_list::timestamptz []) WITH ORDINALITY AS mod(elem, ord) ON vuln.ord = mod.ord
    JOIN vulnerability_data vd ON vd.vulnerability_id = vuln.elem
WHERE vd.last_modified >= mod.elem;

-- name: BatchUpdateVulnerabilityState :exec
WITH root_account AS (
    SELECT COALESCE(
            (
                SELECT root_account_id
                FROM iam_accounts
                WHERE account_id = $1
                LIMIT 1
            ), $1
        ) AS id
),
vuln_list_data AS (
    SELECT vd.vulnerability_data_id
    FROM unnest(@vuln_list::text []) AS vuln_id
        JOIN vulnerability_data vd ON vd.vulnerability_id = vuln_id
),
latest_state_history AS (
    SELECT DISTINCT ON (vuln_data_id) vuln_data_id,
        vulnerability_state
    FROM vulnerability_state_history
    WHERE root_account_id = (
            SELECT id
            FROM root_account
        )
    ORDER BY vuln_data_id,
        state_changed_at DESC
),
insert_active_and_resurfaced AS (
    INSERT INTO vulnerability_state_history (
            vuln_data_id,
            vulnerability_state,
            root_account_id
        )
    SELECT vl.vulnerability_data_id,
        CASE
            WHEN lsh.vulnerability_state = 'New' THEN 'Active'::vulnstate
            WHEN lsh.vulnerability_state = 'Resolved' THEN 'Resurfaced'::vulnstate
        END,
        (
            SELECT id
            FROM root_account
        )
    FROM vuln_list_data vl
        JOIN latest_state_history lsh ON lsh.vuln_data_id = vl.vulnerability_data_id
    WHERE lsh.vulnerability_state = 'New'
        OR lsh.vulnerability_state = 'Resolved'
),
insert_new_and_resolved AS (
    INSERT INTO vulnerability_state_history (
            vuln_data_id,
            vulnerability_state,
            root_account_id
        )
    SELECT COALESCE(
            vl.vulnerability_data_id,
            lsh.vuln_data_id
        ) AS vulnerability_data_id,
        CASE
            WHEN lsh.vuln_data_id IS NULL THEN 'New'::vulnstate
            WHEN vl.vulnerability_data_id IS NULL THEN 'Resolved'::vulnstate
        END AS state_change,
        (
            SELECT id
            FROM root_account
        )
    FROM vuln_list_data vl
        FULL OUTER JOIN latest_state_history lsh ON lsh.vuln_data_id = vl.vulnerability_data_id
    WHERE (
            lsh.vulnerability_state IS NULL
            OR (
                vl.vulnerability_data_id IS NULL
                AND lsh.vulnerability_state != 'Resolved'
            )
        )
)
SELECT 1;

-- name: BatchUpdateVulnerabilityData :exec
UPDATE vulnerability_data
SET vulnerability_name = vuln->>'Name',
    vulnerability_description = vuln->>'Description',
    vulnerability_severity = vuln->>'Severity',
    cvss_score = (vuln->>'CVSSScore')::float,
    created_on = (vuln->>'CreatedOn')::timestamptz,
    last_modified = (vuln->>'LastModified')::timestamptz,
    reference = CASE
        WHEN jsonb_typeof(vuln->'References') = 'array' THEN ARRAY(
            SELECT jsonb_array_elements_text(vuln->'References')
        )
        ELSE ARRAY []::text []
    END
FROM jsonb_array_elements(@vulnerabilities::jsonb) AS vuln
WHERE vulnerability_data.vulnerability_id = vuln->>'VulnerabilityID';


-- name: GetVulnerabilities :many
WITH root_account AS (
    SELECT COALESCE(
            (
                SELECT root_account_id
                FROM iam_accounts
                WHERE account_id = $1
                LIMIT 1
            ), $1
        ) AS id
),
latest_state_history AS (
    SELECT vuln_data_id,
        vulnerability_state
    FROM (
            SELECT *,
                ROW_NUMBER() OVER (
                    PARTITION BY vuln_data_id
                    ORDER BY state_changed_at DESC
                ) AS rn
            FROM vulnerability_state_history
            WHERE root_account_id = (
                    SELECT id
                    FROM root_account
                )
        ) latest
    WHERE rn = 1
)
SELECT *
FROM vulnerability_data
    JOIN latest_state_history lsh ON lsh.vuln_data_id = vulnerability_data.vulnerability_data_id;

-- name: GetVulnerabilitiesStateHistory :many
SELECT *
FROM vulnerability_state_history;

-- name: RetrieveVulnTable :many
WITH root_account AS (
    SELECT COALESCE(
            (
                SELECT root_account_id
                FROM iam_accounts
                WHERE account_id = $1
                LIMIT 1
            ), $1
        ) AS id
),
latest_state_history AS (
    SELECT vuln_data_id,
        vulnerability_state
    FROM vulnerability_state_history
    WHERE root_account_id = (
            SELECT id
            FROM root_account
        )
        AND state_changed_at = (
            SELECT MAX(state_changed_at)
            FROM vulnerability_state_history
            WHERE vuln_data_id = vulnerability_state_history.vuln_data_id
        )
),
last_seen_table AS (
    SELECT vulnerability_id,
        scan_date AS last_seen
    FROM asset_vulnerability_scan avs
    WHERE root_account_id = (
            SELECT id
            FROM root_account
        )
        AND scan_date = (
            SELECT MAX(scan_date)
            FROM asset_vulnerability_scan
            WHERE vulnerability_id = avs.vulnerability_id
        )
),
asset_pairs AS (
    SELECT DISTINCT si.hostname,
        a.asset_id
    FROM asset_vulnerability_scan avs
        JOIN assets a ON a.asset_id = avs.asset_id
        JOIN system_information si ON si.id = a.sysinfo_id
    WHERE a.root_account_id = (
            SELECT id
            FROM root_account
        )
)
SELECT vd.vulnerability_data_id,
    vd.vulnerability_id,
    lsh.vulnerability_state,
    vd.vulnerability_severity,
    vd.cvss_score,
    array_agg(
        ap.hostname
        ORDER BY ap.hostname
    )::TEXT [] AS assets_affected,
    array_agg(
        ap.asset_id
        ORDER BY ap.hostname
    )::UUID [] AS asset_uuids,
    lst.last_seen
FROM vulnerability_data vd
    JOIN latest_state_history lsh ON lsh.vuln_data_id = vd.vulnerability_data_id
    JOIN last_seen_table lst ON lst.vulnerability_id = vd.vulnerability_data_id
    JOIN asset_vulnerability_scan avs ON avs.vulnerability_id = vd.vulnerability_data_id
    JOIN asset_pairs ap ON ap.asset_id = avs.asset_id
GROUP BY vd.vulnerability_data_id,
    vd.vulnerability_id,
    lsh.vulnerability_state,
    vd.vulnerability_severity,
    vd.cvss_score,
    lst.last_seen
ORDER BY vd.cvss_score DESC;

-- name: RetrieveVulnData :one
SELECT vulnerability_name,
    vulnerability_description,
    cvss_score,
    reference,
    created_on,
    last_modified
FROM vulnerability_data
WHERE vulnerability_id = $1;