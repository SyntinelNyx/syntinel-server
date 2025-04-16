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
    AND avs.cve_id NOT IN (
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
    AND avs.cve_id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state = 'resolved';

<< << << < HEAD << << << < HEAD -- name: CalculateNewVulnerabilities :exec
== == == = -- name: CalculateNewVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
>> >> >> > 8e4313b (
    Updated DB Schema
    AND CREATE SQL queries FOR vulnerability scan logic (untested)
) == == == = -- name: CalculateNewVulnerabilities :exec
>> >> >> > 55594c1 (
    Updated DB Schema FOR Scans
    AND Added Relevant Scan Queries
)
SELECT cve_id
FROM unnest($2) AS current_cves(cve_id)
WHERE cve_id NOT IN (
        SELECT avs.cve_id
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
    AND avs.cve_id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

-- name: UpdatePreviouslySeenVulnerabilities :many
WITH current_vulns AS (
    SELECT unnest(@CVE_list::text []) AS cve_id
),
updated AS (
    UPDATE asset_vulnerability_state avs
    SET scan_id = $2,
        vulnerability_state = CASE
            WHEN v.cve_id NOT IN (
                SELECT cve_id
                FROM current_vulns
            )
            AND avs.vulnerability_state != 'Resolved' THEN 'Resolved'
            WHEN v.cve_id IN (
                SELECT cve_id
                FROM current_vulns
            )
            AND avs.vulnerability_state = 'Resolved' THEN 'Resurfaced'
            WHEN v.cve_id IN (
                SELECT cve_id
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
    SELECT cve_id
    FROM current_vulns
    WHERE cve_id NOT IN (
            SELECT cve_id
            FROM vulnerabilities
        )
)
SELECT cve_id::TEXT
FROM new_vulns;


-- name: AddNewVulnerabilities :exec
INSERT INTO vulnerabilities (
        cve_id,
        vulnerability_name,
        vulnerability_description,
        vulnerability_severity,
        cvss_score,
        reference
    )
SELECT vuln->>'CVE_ID',
    vuln->>'VulnerabilityName',
    vuln->>'VulnerabilityDescription',
    vuln->>'VulnerabilitySeverity',
    (vuln->>'CVSSScore')::float,
    -- Handle the References field as an array, default to empty array if not an array
    CASE
        WHEN jsonb_typeof(vuln->'References') = 'array' THEN ARRAY(
            SELECT jsonb_array_elements_text(vuln->'References')
        )
        ELSE ARRAY []::text [] -- Empty array if it's not an array
    END AS reference
FROM jsonb_array_elements(@vulnerabilities::jsonb) AS vuln ON CONFLICT (cve_id) DO NOTHING;

-- name: GetVulnerabilities :many
SELECT cve_id,
    vulnerability_name,
    vulnerability_severity,
    cvss_score
FROM vulnerabilities;