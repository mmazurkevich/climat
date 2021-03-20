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

type MessageFromDevice struct {
	Battery int
	Humidity float32
	Linkquality int
	Pressure int
	Temperature float32
	Voltage int
}

var (
	batteryValue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_battery_value",
		Help: "The total number of processed events",
	})
	temperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_temperature",
		Help: "The total number of processed events",
	})
	humidity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_humidity",
		Help: "The total number of processed events",
	})
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	switch msg.Topic() {
	case "zigbee2mqtt/0x00158d0004850056":
		var message MessageFromDevice
		json.Unmarshal(msg.Payload(), &message)
		batteryValue.Set(float64(message.Battery))
		temperature.Set(float64(message.Temperature))
		humidity.Set(float64(message.Humidity))
		fmt.Printf("Got new temperature value: %f \n", message.Temperature)
	default:
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	var deviceTopic = "zigbee2mqtt/bridge/event"
	var broker = "192.168.0.132"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client, deviceTopic)
	sub(client, "zigbee2mqtt/0x00158d0004850056")

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