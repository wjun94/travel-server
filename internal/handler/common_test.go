package handler

import "testing"

func TestBuildWeatherPayloadReturnsRequestedForecastDays(t *testing.T) {
	response := &amapWeatherResponse{
		Lives: []amapLiveWeather{{
			City:        "北京",
			Weather:     "晴",
			Temperature: "32",
		}},
		Forecasts: []amapForecastWeather{{
			City: "北京",
			Casts: []amapForecastDay{
				{Date: "2026-07-05", Dayweather: "晴", Nightweather: "多云", Daytemp: "35", Nighttemp: "26"},
				{Date: "2026-07-06", Dayweather: "晴", Nightweather: "晴", Daytemp: "36", Nighttemp: "27"},
				{Date: "2026-07-07", Dayweather: "多云", Nightweather: "阴", Daytemp: "34", Nighttemp: "25"},
				{Date: "2026-07-08", Dayweather: "阴", Nightweather: "小雨", Daytemp: "33", Nighttemp: "24"},
			},
		}},
	}

	payload := buildWeatherPayload("北京", 7, response)
	forecast, ok := payload["forecast"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected forecast slice, got %T", payload["forecast"])
	}
	if len(forecast) != 4 {
		t.Fatalf("expected 4 forecast days, got %d", len(forecast))
	}
	if payload["city"] != "北京" {
		t.Fatalf("expected city to be 北京, got %v", payload["city"])
	}
}
