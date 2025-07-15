package source

/* 
url: https://jobicy.com/api/v2/remote-jobs
type: JSON 
options: count, geo, industry, tag
details:
- count max is 100
- ordered most recent to least recent
- rate limited
structure:
- object
- jobs is list at root.jobs
relevant fields:
- url
- jobTitle
- companyName
- jobIndustry
- jobType
- jobGeo
- jobLevel
- pubDate

info: https://jobicy.com/jobs-rss-feed
*/
