## Climate system

### How to start MQTT broker at Raspberry PI?

You should connect to the raspberry via ssh:
    
    ssh pi@192.168.0.132
    password 1

After that we need to start **Mosquitto** server which accept data sent from the device via **ZigBee** protocol and transfer data to MQTT broker

    sudo systemctl enable mosquitto.service
    mosquitto -v

    # Configuration from zigbee2mqtt
    nano /opt/zigbee2mqtt/data/configuration.yaml

    # Start zigbee2mqtt
    cd /opt/zigbee2mqtt
    npm start

### How to get data from MQTT server?

MQTT server URL **mqtt://localhost:1883** backend server use client lib for connecting to MQTT server and subscribe to the topic **zigbee2mqtt/bridge/event** to listen connected devices and afterward subscribe to the device directly and add **temperature data** as a metric for prometheus  

### How to check that server expose metrics?

After the server startup you can check all metrics which application expose to prometheus by the following URL http://localhost:2112/metrics.
At this site we can search metric which we added or which is important for us. 

### How to start and configure prometheus?

    # Start Prometheus.
    # By default, Prometheus stores its database in ./data (flag --storage.tsdb.path).
    ./prometheus --config.file=prometheus.yml
The server will start at default port 9090. At a local machine can open at the browser on http://localhost:9090.

### How to add new metrics?

You should edit `prometheus.yml` file and add the following configuration:
    
    scrape_configs:
    - job_name: myapp
      scrape_interval: 10s
      static_configs:
      - targets:
        - localhost:2112