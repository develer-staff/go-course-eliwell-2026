package main

import "fmt"

type Sensor struct {
	ID       string
	Location string
}

func (s Sensor) Describe() string {
	return fmt.Sprintf("sensor %s at %s", s.ID, s.Location)
}

// Thermostat embeds Sensor: Sensor's fields and methods are promoted to Thermostat.
type Thermostat struct {
	Sensor
	TargetTemp float64
}

func main() {
	t := Thermostat{
		Sensor: Sensor{
			ID:       "TH-101",
			Location: "Cold Room A",
		},
		TargetTemp: -18.5,
	}

	// Promoted fields and methods can be accessed directly on Thermostat:
	fmt.Println("ID:", t.ID)
	fmt.Println("Location:", t.Location)
	fmt.Println("Describe:", t.Describe())
	fmt.Printf("Target: %.1f°C\n", t.TargetTemp)

	// Note: The embedded struct is still accessible as a named field:
	fmt.Println("Inner struct:", t.Sensor.Describe())
}
