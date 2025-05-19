package weather

// swagger:model Weather
type Weather struct {

	// Current temperature
	Temperature float64 `json:"temperature"`

	// Current humidity percentage
	Humidity float64 `json:"humidity"`

	// Weather description
	Description string `json:"description"`
}

type WeatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  float64 `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}
