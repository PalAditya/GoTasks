package models

type Response struct {
	Success          bool `json:"success"`
	Data             Data
	LastRefreshed    string `json:"lastRefreshed"`
	LastOriginUpdate string `json:"lastOriginUpdate"`
}

type Data struct {
	Regional []Regional
	Summary  Summary
	X        map[string]interface{} `json:"-"`
}

type Summary struct {
	Indiancases int64                  `json:"confirmedCasesIndian"`
	Discharged  int64                  `json:"discharged"`
	X           map[string]interface{} `json:"-"`
}

type ApiResponse struct {
	LatestRecord string `json:"latestRecord"`
	Inserted     bool   `json:"inserted"`
	Modified     bool   `json:"modified"`
}
