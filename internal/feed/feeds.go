package feed

var FEEDS = []Feed{
	{
		Name: "Remotive",
		URL:  "http://localhost:8000/remotive.rss",
		Mapping: ItemMap{
			Company:  "company",
			Location: "location",
			Kind:     "type",
		},
	},
}
