-- name: CalculateResolvedVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.cve_id NOT IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

-- name: CalculateResurfacedVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.cve_id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state = 'resolved';

-- name: CalculateNewVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
SELECT cve_id
FROM unnest($2) AS current_cves(cve_id)
WHERE cve_id NOT IN (
        SELECT avs.cve_id
        FROM asset_vulnerability_state avs
            JOIN assets a ON a.asset_id = avs.asset_id
            JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
        WHERE asset_id = $1;
);

-- name: CalculateNotAffectedVulnerabilities Role :exec
-- $1: asset_id - asset the scan was intiated on
-- $2: current_cve_list - list of cves returned by most recent scan
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.cve_id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

-- name: UpdatePreviouslySeenVulnerabilitiesOnAsset
-- $1: asset_id - asset the scan was initiated on
-- $2: scan_id - relevant scan
-- $3: current_cve_list - list of cves returned by most recent scan
WITH current_vulns AS (
    SELECT unnest($3) AS cve_id
)
UPDATE asset_vulnerability_state avs
SET scan_id = $2,
    vulnerability_state = CASE
        WHEN avs.cve_id NOT IN (
            SELECT cve_id
            FROM current_vulns
        )
        AND avs.vulnerability_state != 'resolved' THEN 'resolved'
        WHEN avs.cve_id IN (
            SELECT cve_id
            FROM current_vulns
        )
        AND avs.vulnerability_state = 'resolved' THEN 'resurfaced'
        ELSE avs.vulnerability_state
    END
WHERE avs.asset_id = $1
    AND avs.cve_id IN (
        SELECT cve_id
        FROM current_vulns
    );