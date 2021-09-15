package apis

//structs for the /data/:lat/:long endpoint
type GeoResponse struct {
	X       map[string]interface{} `json:"-"`
	Address Address
}

type Address struct {
	State string                 `json:"state"`
	X     map[string]interface{} `json:"-"`
}

type StateNotFound struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

type MongoResponse struct {
	Region      []Regional `bson:"res"`
	LastUpdated string     `bson:"lastUpdated"`
	ModifiedAt  string     `bson:"modifiedAt"`
	RecordDate  string     `bson:"recordDate"`
	Summary     DBSummary  `bson:"summary"`
}

type UserResponse struct {
	State             string `json:"state"`
	TotalCasesByState int64  `json:"totalCasesByState"`
	TotalCasesInIndia int64  `json:"totalCasesInIndia"`
	LastUpdated       string `json:"lastUpdated"`
}

//Structs for the /api endpoint
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

type DBSummary struct {
	Total      int64 `json:"total"`
	Discharged int64 `json:"discharged"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type Summary struct {
	Indiancases int64                  `json:"confirmedCasesIndian"`
	Discharged  int64                  `json:"discharged"`
	X           map[string]interface{} `json:"-"`
}

type Regional struct {
	Loc           string `json:"loc"`
	Indiancases   int64  `json:"confirmedCasesIndian"`
	Foreigncases  int64  `json:"confirmedCasesForeign"`
	Discharged    int64  `json:"discharged"`
	Deaths        int64  `json:"deaths"`
	Totalconfimed int64  `json:"totalConfirmed"`
}