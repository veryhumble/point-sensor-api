package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	i2c "github.com/d2r2/go-i2c"
	si7021 "github.com/d2r2/go-si7021"
	"github.com/go-ini/ini"
	"github.com/julienschmidt/httprouter"
)

var SensorId = 1000
var SensorLocation = ""
var SensorHardwareVersion = "1.0"
var SensorHardware = ""
var SensorType = ""

// Index returns general information about this sensor
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	sensor := Sensor{
		ID:       SensorId,
		Name:     SensorLocation,
		Version:  SensorHardwareVersion,
		Hardware: SensorHardware,
		Type:     SensorType,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sensor)
}

// Point returns a measurement
func Point(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Create new connection to i2c-bus on 1 line with address 0x40.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x40, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	sensor := si7021.NewSi7021()

	rh, temp, err := sensor.ReadRelativeHumidityAndTemperature(i2c)
	if err != nil {
		log.Fatal(err)
	}

	measurement := Measure{
		ID:               SensorId,
		Name:             SensorLocation,
		Type:             SensorType,
		Timestamp:        makeTimestamp(),
		Temperature:      temp,
		RelativeHumidity: rh,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(measurement)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func main() {
	fmt.Println("Starting Point Sensor API...")
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	SensorLocation = cfg.Section("app").Key("sensor_location").String()
	SensorType = cfg.Section("app").Key("sensor_type").String()
	SensorHardware = cfg.Section("app").Key("sensor_hardware").String()
	SensorHardwareVersion = cfg.Section("app").Key("sensor_hardware_version").String()
	SensorId = cfg.Section("app").Key("sensor_id").MustInt()
	var port = cfg.Section("server").Key("http_port").String()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/point", Point)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// Sensor Information
type Sensor struct {
	Name     string /**< sensor name */
	Version  string /**< version of the hardware + driver */
	ID       int    /**< unique sensor identifier */
	Hardware string /**< the sensor hardware name */
	Type     string /**< this sensor's type (ex. SENSOR_TYPE_LIGHT) */
	// MaxValue   float32 /**< maximum value of this sensor's value in SI units */
	// MinValue   float32 /**< minimum value of this sensor's value in SI units */
	// Resolution float32 /**< smallest difference between two values reported by this sensor */
	// MinDelay   int     /**< min delay in microseconds between events. zero = not a constant rate */
}

// Measure represents a measurement
type Measure struct {
	Name             string
	ID               int
	Type             string
	Timestamp        int64
	Temperature      float32 /**< temperature is in degrees centigrade (Celsius) */
	Distance         float32 /**< distance in centimeters */
	Light            float32 /**< light in SI lux units */
	Pressure         float32 /**< pressure in hectopascal (hPa) */
	RelativeHumidity float32 /**< relative humidity in percent */
	Current          float32 /**< current in milliamps (mA) */
	Voltage          float32 /**< voltage in volts (V) */
}
