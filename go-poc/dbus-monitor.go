// https://github.com/SCHKN/systemd-monitoring/blob/master/main.go

package main

import (
	"fmt"
  //"encoding/json"
	// "log"
	// "net/url"
	// "os"
	// "time"
	// "strings"
  //"github.com/coreos/go-systemd/dbus"
	"github.com/godbus/dbus"
	"os"

	//client "github.com/influxdata/influxdb1-client"
)

func main() {

	fmt.Printf("Battery Level: ")
	parseBatLevel()
  fmt.Printf("\n")

	// Get all registered names
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to bus:", err)
		os.Exit(1)
	}

	var s []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of owned names:", err)
		os.Exit(1)
	}

	fmt.Println("Currently owned names on the session bus:")
	for _, v := range s {
		fmt.Println(v)
	}


	// Send notification to the Notifications server
	conn1, err1 := dbus.SessionBus()
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to bus:", err1)
		os.Exit(1)
	}
	obj := conn1.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, "", uint32(0),
		"", "Test", "This is a test of the DBus bindings for go.", []string{},
		map[string]dbus.Variant{}, int32(5000))
	if call.Err != nil {
		panic(call.Err)
	}


	// Listen to all signals
	conn2, err2 := dbus.SystemBus() // <------ This distinguishes between system and session buses
	if err2 != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to bus:", err2)
		os.Exit(1)
	}

	for _, v := range []string{"method_call", "method_return", "error", "signal"} {
		call := conn2.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			"eavesdrop='false',type='"+v+"'")  // <-------- eavesdrop is deprecatred and does not work on system bus
		if call.Err != nil {
			fmt.Fprintln(os.Stderr, "Failed to add match:", call.Err)
			os.Exit(1)
		}
	}
	c := make(chan *dbus.Message, 10)
	conn2.Eavesdrop(c)
	fmt.Println("Listening for everything")
	for v := range c {
		fmt.Println(v)
	}

}


// parseBatLevel connects to the system bus and get the State and Percentage
// properties from the UPower's BAT object. It returns the level in percents
// and integer status, which means:
//
//  0: Unknown
//  1: Charging
//  2: Discharging
//  3: Empty
//  4: Fully charged
//  5: Pending charge
//  6: Pending discharge
//
func parseBatLevel() { // (int, uint32) {
	conn, _ := dbus.SystemBus()
	pth := fmt.Sprintf("/org/freedesktop/UPower/devices/battery_BAT0")
	object := conn.Object(
		"org.freedesktop.UPower",
		dbus.ObjectPath(pth),
	)
	lvl, _ := object.GetProperty("org.freedesktop.UPower.Device.Percentage")
	state, _ := object.GetProperty("org.freedesktop.UPower.Device.State")

	fmt.Printf("%f\n", lvl.Value().(float64))
	fmt.Printf("%d\n", state.Value().(uint32))


	//return int(lvl.Value().(float64)), state.Value().(uint32)
}



//   // var influxClient *client.Client = initInflux()
//
// 	conn, err := dbus.New()
// 	// Ensuring that the connection is closed at some point.
// 	defer conn.Close()
// 	if err != nil {
// 		fmt.Println("Could not create a new connection object.")
// 		return
// 	}
//
// 	// Subscribing to systemd dbus events.
// 	err = conn.Subscribe()
// 	if err != nil {
// 		fmt.Println("Could not subscribe to the bus.")
// 		return
// 	}
//
// 	updateCh := make(chan *dbus.PropertiesUpdate, 256)
// 	errCh := make(chan error, 256)
//
// 	// Properties (signals here) changes will be saved to those objects.
// 	conn.SetPropertiesSubscriber(updateCh, errCh)
//
// 	for {
// 		select {
// 		case update := <-updateCh:
//
//       print_dbus_msg(update)
//       //var messages []dbusMessage = getMessage(update)
// 			// var points []client.Point = getPoint(update)
//
// 			// influxClient.Write(client.BatchPoints{
// 			// 	Points:           points,
// 			// 	Database:         "systemd",
// 			// 	RetentionPolicy:  "autogen",
// 			// 	Precision:        "ms",
// 			// 	WriteConsistency: "any",
// 			// })
//
// 		case err := <-errCh:
// 			fmt.Println(err)
// 		}
// 	}
// }
//
// type dbusMessage struct {
//     time  string
//     sender  string
// }
//
// func print_dbus_msg(update *dbus.PropertiesUpdate) {
//
//     fmt.Println("-----------> New Message: ")
//     fmt.Println("UnitName: ", update.UnitName)
//     fmt.Println("Changed: \n")
//
//     for k, v := range update.Changed {
// 			if(!strings.Contains(k, "Timestamp") ) {
//         fmt.Printf(k, "    :   ", v.String(), "\n\n")
// 			}
//     }
//
// }





// func getMessage(properties *dbus.PropertiesUpdate) []dbusMessage {
//
// 	activeState := properties.Changed["ActiveState"].String()
//
// 	var point client.Point = client.Point{
// 		Measurement: "services",
// 		Tags: map[string]string{
// 			"service": properties.UnitName,
// 		},
// 		Fields: map[string]interface{}{
// 			"state": activeState,
// 			"value": getStateValue(activeState),
// 		},
// 		// Need to use the timestamp provided by dbus..
// 		Time:      time.Now(),
// 		Precision: "ms",
// 	}
//
// 	points := [1]client.Point{point}
//
// 	return points[0:1]
//
// }
// func getPoint(properties *dbus.PropertiesUpdate) []client.Point {
//
// 	activeState := properties.Changed["ActiveState"].String()
//
// 	var point client.Point = client.Point{
// 		Measurement: "services",
// 		Tags: map[string]string{
// 			"service": properties.UnitName,
// 		},
// 		Fields: map[string]interface{}{
// 			"state": activeState,
// 			"value": getStateValue(activeState),
// 		},
// 		// Need to use the timestamp provided by dbus..
// 		Time:      time.Now(),
// 		Precision: "ms",
// 	}
//
// 	points := [1]client.Point{point}
//
// 	return points[0:1]
//
// }

// func initInflux() *client.Client {
//
// 	host, err := url.Parse(fmt.Sprintf("http://%s:%d", "localhost", 8086))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	conf := client.Config{
// 		URL:      *host,
// 		Username: os.Getenv("admin"),
// 		Password: os.Getenv("admin"),
// 	}
//
// 	influxConnection, err := client.NewClient(conf)
//
// 	if err != nil {
// 		fmt.Println("Error building a new client for InfluxDB.")
// 	}
//
// 	return influxConnection
// }

// // active (1), reloading, inactive (0), failed (-1), activating, deactivating
// func getStateValue(state string) int {
//
// 	switch state {
// 	case "\"failed\"":
// 		return -1
// 	case "\"inactive\"":
// 		return 0
// 	case "\"active\"":
// 		return 1
// 	case "\"reloading\"":
// 		return 2
// 	case "\"activating\"":
// 		return 4
// 	case "\"deactivating\"":
// 		return 5
// 	default:
// 		panic("Unhandled dbus active state value")
// 	}
// }
