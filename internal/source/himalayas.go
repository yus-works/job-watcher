package feed

/* 
url: https://himalayas.app/jobs/api
type: JSON 
options: offset, limit
details:
- limit max is 20
- ordered most recent to least recent
- offset is offset not page number
- rate limited
structure:
- object
- jobs is list at root.jobs
relevant fields:
- title
- companyName
- employmentType
- seniority (list)
- locationRestrictions
- timezoneRestrictions
- categories (list)
- pubDate
- expiryDate
- applicationLink
info: https://himalayas.app/api
*/
