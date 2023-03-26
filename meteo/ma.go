package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Message example:
// {"battery":100,"humidity":22.01,"linkquality":89,"pressure":1018,"temperature":24.93,"voltage":3115}
type MessageFromDevice struct {
	Battery     int
	Humidity    float32
	Linkquality int
	Pressure    int
	Temperature float32
	Voltage     int
}

// Message example:
// {"data":{"friendly_name":"0x00158d0004850056","ieee_address":"0x00158d0004850056"},"type":"device_announce"}
type DeviceAnnounceMessage struct {
	Data DeviceInfo
	Type string
}

type DeviceInfo struct {
	FriendlyName string
	IeeeAddress  string
}

var (
	deviceSet    = make(map[string]bool)
	deviceTopic  = "zigbee2mqtt/bridge/event"
	broker       = "192.168.10.13"
	port         = 1883
	mqttClient   mqtt.Client
	batteryValue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "xiaomi_device_battery_value",
		Help: "The total number of processed events",
	})
	temperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "xiaomi_device_temperature",
		Help: "The total number of processed events",
	})
	humidity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "xiaomi_device_humidity",
		Help: "The total number of processed events",
	})
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	topic := msg.Topic()
	if deviceSet[topic] {
		var message MessageFromDevice
		json.Unmarshal(msg.Payload(), &message)
		batteryValue.Set(float64(message.Battery))
		temperature.Set(float64(message.Temperature))
		humidity.Set(float64(message.Humidity))
		fmt.Printf("Got new temperature value: %f \n", message.Temperature)
	} else if topic == deviceTopic {
		var message DeviceAnnounceMessage
		json.Unmarshal(msg.Payload(), &message)
		if message.Type == "device_announce" && !deviceSet[message.Data.IeeeAddress] {
			var deviceSubscriptionTopic = "zigbee2mqtt/" + message.Data.IeeeAddress
			deviceSet[deviceSubscriptionTopic] = true
			sub(mqttClient, deviceSubscriptionTopic)
		}
	} else {
		fmt.Printf("Default handler for unknown topic")
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("mqtt_client")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(mqttClient, deviceTopic)
	subToDefaultDevise(mqttClient)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
	select {}
	//client.Disconnect(250)
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s \n", topic)
}

func subToDefaultDevise(client mqtt.Client) {
	deviceId := "zigbee2mqtt/0x00158d0004850056"
	sub(client, deviceId)
	deviceSet[deviceId] = true
}
