-- name: CalculateResolvedVulnerabilities :exec
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.cve_id NOT IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

-- name: CalculateResurfacedVulnerabilities :exec
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.cve_id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state = 'resolved';

-- name: CalculateNewVulnerabilities :exec
SELECT cve_id
FROM unnest($2) AS current_cves(cve_id)
WHERE cve_id NOT IN (
        SELECT avs.cve_id
        FROM asset_vulnerability_state avs
            JOIN assets a ON a.asset_id = avs.asset_id
            JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
        WHERE avs.asset_id = $1
    );

-- name: CalculateNotAffectedVulnerabilities :exec
SELECT avs.vulnerability_id
FROM asset_vulnerability_state avs
    JOIN assets a ON a.asset_id = avs.asset_id
    JOIN vulnerabilities v ON v.vulnerability_id = avs.vulnerability_id
WHERE a.asset_id = $1
    AND avs.cve_id IN (
        SELECT unnest($2)
    )
    AND vulnerability_state != 'resolved';

-- name: UpdatePreviouslySeenVulnerabilities :exec
WITH current_vulns AS (
    SELECT unnest(@CVE_list::text []) AS cve_id
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