package postgres

var insertQuery = `
    INSERT INTO 
    findings(finding_id, detection_name, collision_slug, first_event, last_event, raw_events) 
    values ($1, $2, $3, $4, $5, $6)
`

var mergeQuery = `
    UPDATE findings 
    SET raw_events = array_cat(raw_events, $1)
    WHERE collision_slug = $2
    AND first_event = $3
    AND last_event = $4;
`
