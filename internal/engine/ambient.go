package engine

import (
	"math"
	"time"
)

type RGBColor struct {
	R, G, B int
}

type Keyframe struct {
	Hour      float64
	SkyTop    RGBColor
	SkyBottom RGBColor
}

type AmbientState struct {
	IsDaytime      bool
	SunX, SunY     int
	MoonX, MoonY   int
	SkyTop         RGBColor
	SkyBottom      RGBColor
	MoonPhase      float64
	MoonPhaseName  string
}

// 14 Color keyframes across the 24-hour cycle.
var SkyKeyframes = []Keyframe{
	{Hour: 0.0, SkyTop: RGBColor{11, 12, 16}, SkyBottom: RGBColor{31, 40, 51}},    // Deep Night
	{Hour: 3.0, SkyTop: RGBColor{15, 16, 22}, SkyBottom: RGBColor{21, 27, 38}},    // Pre-dawn
	{Hour: 5.0, SkyTop: RGBColor{44, 62, 80}, SkyBottom: RGBColor{253, 116, 108}}, // Dawn Twilight
	{Hour: 6.0, SkyTop: RGBColor{230, 92, 0}, SkyBottom: RGBColor{249, 212, 35}},  // Sunrise
	{Hour: 7.5, SkyTop: RGBColor{33, 147, 176}, SkyBottom: RGBColor{109, 213, 237}}, // Early Morning
	{Hour: 9.0, SkyTop: RGBColor{43, 88, 118}, SkyBottom: RGBColor{78, 67, 118}},   // Morning
	{Hour: 12.0, SkyTop: RGBColor{26, 41, 128}, SkyBottom: RGBColor{38, 208, 206}}, // Solar Noon
	{Hour: 15.0, SkyTop: RGBColor{243, 144, 79}, SkyBottom: RGBColor{59, 67, 113}}, // Late Afternoon
	{Hour: 17.0, SkyTop: RGBColor{241, 101, 41}, SkyBottom: RGBColor{228, 77, 38}}, // Golden Hour
	{Hour: 18.0, SkyTop: RGBColor{233, 100, 67}, SkyBottom: RGBColor{144, 78, 149}}, // Sunset
	{Hour: 19.0, SkyTop: RGBColor{44, 62, 80}, SkyBottom: RGBColor{52, 152, 219}},  // Dusk
	{Hour: 20.0, SkyTop: RGBColor{0, 4, 40}, SkyBottom: RGBColor{0, 78, 146}},      // Late Dusk
	{Hour: 22.0, SkyTop: RGBColor{15, 32, 39}, SkyBottom: RGBColor{32, 58, 67}},    // Night
	{Hour: 24.0, SkyTop: RGBColor{11, 12, 16}, SkyBottom: RGBColor{31, 40, 51}},    // Midnight Loop
}

type Star struct {
	X, Y  int
	Phase int // twinkle timing offset
}

// GenerateStars deterministically generates a set of stars based on grid dimensions.
func GenerateStars(width, height int, count int) []Star {
	stars := make([]Star, 0, count)
	// We want stars scattered in the sky (upper section of screen, say Y from 2 to 14)
	if height < 15 {
		return stars
	}

	for i := 0; i < count; i++ {
		// Use a pseudo-random hash function to place stars deterministically
		seed := int64(i * 12345)
		x := int((seed ^ 987654) % int64(width-10)) + 5
		y := int((seed ^ 345678) % 11) + 2 // range 2 to 12
		phase := int(seed % 3)
		stars = append(stars, Star{X: x, Y: y, Phase: phase})
	}
	return stars
}

// GetMoonPhase returns the current lunar phase fraction (0.0 - 1.0) and description.
func GetMoonPhase(t time.Time) (float64, string) {
	knownNewMoon := time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)
	diff := t.Sub(knownNewMoon).Hours() / 24.0
	cycles := diff / 29.530588853
	phase := cycles - math.Floor(cycles)

	var name string
	if phase < 0.03 || phase >= 0.97 {
		name = "New Moon"
	} else if phase < 0.22 {
		name = "Waxing Crescent"
	} else if phase < 0.28 {
		name = "First Quarter"
	} else if phase < 0.47 {
		name = "Waxing Gibbous"
	} else if phase < 0.53 {
		name = "Full Moon"
	} else if phase < 0.72 {
		name = "Waning Gibbous"
	} else if phase < 0.78 {
		name = "Third Quarter"
	} else {
		name = "Waning Crescent"
	}

	return phase, name
}

// InterpolateColor calculates the color between c1 and c2 at percentage t.
func InterpolateColor(c1, c2 RGBColor, t float64) RGBColor {
	if t <= 0.0 {
		return c1
	}
	if t >= 1.0 {
		return c2
	}
	return RGBColor{
		R: int(math.Round(float64(c1.R) + t*float64(c2.R-c1.R))),
		G: int(math.Round(float64(c1.G) + t*float64(c2.G-c1.G))),
		B: int(math.Round(float64(c1.B) + t*float64(c2.B-c1.B))),
	}
}

// GetAmbientState calculates celestial coordinates and colors for a given time of day.
func GetAmbientState(t time.Time, sunriseHour, sunsetHour int) AmbientState {
	if sunriseHour <= 0 {
		sunriseHour = 7
	}
	if sunsetHour <= 0 {
		sunsetHour = 19
	}

	hour := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0

	// Interpolate Sky Colors
	var skyTop, skyBottom RGBColor
	for i := 0; i < len(SkyKeyframes)-1; i++ {
		kf1 := SkyKeyframes[i]
		kf2 := SkyKeyframes[i+1]
		if hour >= kf1.Hour && hour <= kf2.Hour {
			denom := kf2.Hour - kf1.Hour
			progress := 0.0
			if denom > 0 {
				progress = (hour - kf1.Hour) / denom
			}
			skyTop = InterpolateColor(kf1.SkyTop, kf2.SkyTop, progress)
			skyBottom = InterpolateColor(kf1.SkyBottom, kf2.SkyBottom, progress)
			break
		}
	}

	isDaytime := hour >= float64(sunriseHour) && hour < float64(sunsetHour)
	sunX, sunY := -1, -1
	moonX, moonY := -1, -1

	if isDaytime {
		// Sun Orbit Math
		progress := (hour - float64(sunriseHour)) / float64(sunsetHour-sunriseHour)
		sunX = int(10.0 + 80.0*progress)
		elevation := math.Sin(progress * math.Pi)
		sunY = int(16.0 - 13.0*elevation)
	} else {
		// Moon Orbit Math
		nightDuration := 24.0 - float64(sunsetHour-sunriseHour)
		var progress float64
		if hour >= float64(sunsetHour) {
			progress = (hour - float64(sunsetHour)) / nightDuration
		} else {
			progress = (hour + (24.0 - float64(sunsetHour))) / nightDuration
		}
		moonX = int(10.0 + 80.0*progress)
		elevation := math.Sin(progress * math.Pi)
		moonY = int(16.0 - 13.0*elevation)
	}

	phaseFraction, phaseName := GetMoonPhase(t)

	return AmbientState{
		IsDaytime:     isDaytime,
		SunX:          sunX,
		SunY:          sunY,
		MoonX:         moonX,
		MoonY:         moonY,
		SkyTop:        skyTop,
		SkyBottom:     skyBottom,
		MoonPhase:     phaseFraction,
		MoonPhaseName: phaseName,
	}
}
