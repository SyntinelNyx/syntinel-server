<< << << < HEAD << << << < HEAD -- name: CalculateResolvedVulnerabilities :exec
== == == = -- name: CalculateResolvedVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
>> >> >> > 8e4313b (
    Updated DB Schema
    AND CREATE SQL queries FOR vulnerability scan logic (untested)
) == == == = -- name: CalculateResolvedVulnerabilities :exec
>> >> >> > 55594c1 (
    Updated DB Schema FOR Scans
    AND Added Relevant Scan Queries
)
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.id NOT IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

<< << << < HEAD << << << < HEAD -- name: CalculateResurfacedVulnerabilities :exec
== == == = -- name: CalculateResurfacedVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
>> >> >> > 8e4313b (
    Updated DB Schema
    AND CREATE SQL queries FOR vulnerability scan logic (untested)
) == == == = -- name: CalculateResurfacedVulnerabilities :exec
>> >> >> > 55594c1 (
    Updated DB Schema FOR Scans
    AND Added Relevant Scan Queries
)
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state = 'resolved';

-- name: CalculateNewVulnerabilities :exec
SELECT id
FROM unnest($2) AS current_vulns(id)
WHERE id NOT IN (
        SELECT avs.id
        FROM asset_vulnerability_state avs
            JOIN assets a ON a.asset_id = avs.asset_id << << << < HEAD
            JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id << << << < HEAD
        WHERE avs.asset_id = $1
    );

-- name: CalculateNotAffectedVulnerabilities :exec
== == == =
WHERE asset_id = $1;
);

-- name: CalculateNotAffectedVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
>> >> >> > 8e4313b (
    Updated DB Schema
    AND CREATE SQL queries FOR vulnerability scan logic (untested)
) == == == =
JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE avs.asset_id = $1
);

-- name: CalculateNotAffectedVulnerabilities :exec
>> >> >> > 55594c1 (
    Updated DB Schema FOR Scans
    AND Added Relevant Scan Queries
)
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

-- name: UpdatePreviouslySeenVulnerabilities :many
WITH current_vulns AS (
    SELECT unnest(@vuln_list::text []) AS id
),
updated AS (
    UPDATE asset_vulnerability_state avs
    SET scan_id = $2,
        vulnerability_state = CASE
            WHEN v.id NOT IN (
                SELECT id
                FROM current_vulns
            )
            AND avs.vulnerability_state != 'Resolved' THEN 'Resolved'
            WHEN v.id IN (
                SELECT id
                FROM current_vulns
            )
            AND avs.vulnerability_state = 'Resolved' THEN 'Resurfaced'
            WHEN v.id IN (
                SELECT id
                FROM current_vulns
            )
            AND avs.vulnerability_state = 'New' THEN 'Active'
            ELSE avs.vulnerability_state
        END
    FROM vulnerabilities v
    WHERE avs.asset_id = $1
        AND avs.vulnerability_id = v.vulnerability_id
),
new_vulns AS (
    SELECT id
    FROM current_vulns
    WHERE id NOT IN (
            SELECT id
            FROM vulnerabilities
        )
)
SELECT id::TEXT
FROM new_vulns;

-- name: PrepareVulnerabilityState :exec
UPDATE vulnerabilities
SET vulnerability_state = 'Active'
WHERE vulnerability_state = 'New';


-- name: BatchUpdateVulnerabilityState :exec
UPDATE vulnerabilities v
SET vulnerability_state = CASE
        WHEN vl.vulnerability_id IS NULL
        AND v.vulnerability_state != 'Resolved' THEN 'Resolved'
        WHEN vl.vulnerability_id IS NOT NULL
        AND v.vulnerability_state = 'Resolved' THEN 'Resurfaced'
        ELSE v.vulnerability_state
    END
FROM (
        SELECT unnest(@vuln_list::text []) AS vulnerability_id
    ) AS vl
WHERE v.vulnerability_id = vl.vulnerability_id
    OR v.vulnerability_state != 'Resolved';

-- name: InsertNewVulnerabilities :exec
INSERT INTO vulnerabilities(vulnerability_id)
SELECT vulnerability_id
FROM unnest(@vuln_list::text []) AS vulnerability_id
WHERE NOT EXISTS (
        SELECT 1
        FROM vulnerabilities v
        WHERE v.vulnerability_id = vulnerability_id
    );

-- name: RetrieveUnchangedVulnerabilities :many
SELECT v.vulnerability_id
FROM unnest(@vuln_list::text []) WITH ORDINALITY AS vuln(elem, ord)
    JOIN unnest(@modified_list::timestamptz []) WITH ORDINALITY AS mod(elem, ord) ON vuln.ord = mod.ord
    JOIN vulnerabilities v ON v.vulnerability_id = vuln.elem
WHERE v.last_modified >= mod.elem;


-- name: BatchUpdateVulnerabilityData :exec
UPDATE vulnerabilities
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
WHERE vulnerabilities.vulnerability_id = vuln->>'ID';



-- name: GetVulnerabilities :many
SELECT *
FROM vulnerabilities;