package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	speedtest "github.com/showwin/speedtest-go/speedtest"
)

type HomeAssistantConfig struct {
	Name     string `json:"name"`                // sensor-name in home-assistant
	Unit     string `json:"unit_of_measurement"` // unit of the sensor
	Topic    string `json:"state_topic"`         // MQTT-Topic where home-assistant listens for state updates
	Template string `json:"value_template"`      // jinja2 template how to parse the sensor value
}

// Retrieves a setting from OS environment variables. If the env is not provided
// the fallback string is used.
// If the fallback string is empty, a panic is raised.
func getSettings(key string, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		if len(fallback) == 0 {
			panic(errors.New(key + ": missing value"))
		}

		return fallback
	}
	return value
}

// Registers three sensors in home-assistant using its MQTT discovery feature.
// More information can be found here: https://www.home-assistant.io/docs/mqtt/discovery/
func registerHomeAssistantSensors(client mqtt.Client, discoveryTopic string, topic string, name string) {
	// Each sensor needs a own config message.
	payloads := [3]HomeAssistantConfig{
		{Name: "%s-ping", Unit: "ms", Topic: "%s", Template: "{{ value_json.latency }} "},
		{Name: "%s-download", Unit: "Mbps", Topic: "%s", Template: "{{ value_json.dl_speed }} "},
		{Name: "%s-upload", Unit: "Mbps", Topic: "%s", Template: "{{ value_json.ul_speed }} "},
	}

	for i := range payloads {
		// Before sending, the "sensor" name and the base-topic is inserted.
		payloads[i].Name = fmt.Sprintf(payloads[i].Name, name)
		payloads[i].Topic = fmt.Sprintf(payloads[i].Topic, topic)

		// The Entity-Type in Home-Assistant is set given the topic of the configuration message
		topic := discoveryTopic + "/sensor/" + payloads[i].Name + "/config"
		payload, _ := json.Marshal(payloads[i])
		// The discovery messages are retained
		client.Publish(topic, 0, true, payload)
	}
}

// Starts the ping, download and upload test.
// Basically the sample from the speedtest-go Readme:
// https://github.com/showwin/speedtest-go#api-usage
func runSpeedTest() (*speedtest.Server, error) {
	user, _ := speedtest.FetchUserInfo()

	serverList, _ := speedtest.FetchServerList(user)
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		s.PingTest()
		s.DownloadTest(false)
		s.UploadTest(false)

		return s, nil
	}
	return nil, errors.New("failed to acquire a server")
}

// Publishes a MQTT Message to the provided topic. The complete speedtest.Server
// struct is published.
func publishSpeedTest(client mqtt.Client, topic string, server *speedtest.Server) {
	if server.CheckResultValid() {
		panic(errors.New("speedtest result is invalid"))
	}

	message, _ := json.Marshal(server)
	if token := client.Publish(topic, 0, false, string(message)); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func main() {

	// Essential ENV configurations. These are required to start.
	// MQTT_BROKER: URL of your broker. e.g. tcp://127.0.0.1:1883
	broker := getSettings("MQTT_BROKER", "")
	// MQTT_USERNAME and MQTT_PASSWORD: Your MQTT credentials
	username := getSettings("MQTT_USERNAME", "")
	password := getSettings("MQTT_PASSWORD", "")

	// MQTT_TOPIC: Base topic for the state messages.
	// If you want to use hierarchy topic, you should update MQTT_NAME as well.
	topic := getSettings("MQTT_TOPIC", "speedtest")

	// Home-Assistant discovery base-topic. Set this to match your home-assistant setup
	discovery := getSettings("MQTT_HOME_ASSISTANT_DISCOVERY", "homeassistant")
	// Name of the home-assistant sensors
	name := getSettings("MQTT_NAME", topic)

	logLevel := getSettings("MQTT_LOG_LEVEL", "INFO")

	switch logLevel {
	case "DEBUG":
		mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
		fallthrough
	case "INFO":
		mqtt.WARN = log.New(os.Stdout, "[INFO]  ", 0)
		fallthrough
	case "ERROR":
		fallthrough
	default:
		mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
		mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(password)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if len(discovery) > 0 {
		registerHomeAssistantSensors(client, discovery, topic, name)
	}

	result, err := runSpeedTest()
	if err != nil {
		panic(err)
	}
	publishSpeedTest(client, topic, result)
}
