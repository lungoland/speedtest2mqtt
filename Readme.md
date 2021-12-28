# speedtest2mqtt
Measures internet speed and publishes the results to an MQTT broker.

Based on
- https://github.com/showwin/speedtest-go and
- https://github.com/eclipse/paho.mqtt.golang


## Usage
The application is configured through environment variables. The essential ones are:

| | | |
|-|-|-|
|MQTT_BROKER| URL to your broker|tcp://127.0.0.1:1883|
|MQTT_USERNAME||
|MQTT_PASSWORD||
|MQTT_TOPIC|Publish topic of the results|speedtest|
|MQTT_HOME_ASSISTANT_DISCOVERY|Discovery topic for home-assistant|homeassistant|
|MQTT_NAME|Sensor name in home-assistant|speedtest|


The application will perform a single measurement and publish the result into the configured topic and exit. It is intented to be triggered by an k8s [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/).

All output is written to stdout.

## Docker
Multi-arch images ca be build using docker buildx.

```
docker buildx build --platform linux/arm/v7,linux/arm64,linux/amd64 --tag speedtest2mqtt:latest .
```