package controllers

import (
	"billingo/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

type MqttPublisher struct {
	name       string
	manager    *Manager
	client     mqtt.Client
	topic      string
	brokerURL  string
	username   string
	password   string
	retryDelay time.Duration
	db         *gorm.DB
}

func NewMQTTPublisherTask(name, brokerURL, topic, username, password string) *MqttPublisher {
	return &MqttPublisher{
		name:       name,
		topic:      topic,
		brokerURL:  brokerURL,
		username:   username,
		password:   password,
		retryDelay: 5 * time.Second, // Retry every 5 seconds
	}
}

func (m *MqttPublisher) Setup(db *gorm.DB, manager *Manager) {
	m.manager = manager
	m.db = db
}

func (m *MqttPublisher) String() string {
	return m.name
}

// Main function to handle the MQTT publishing task.
func (m *MqttPublisher) Main(ctx context.Context) {
	// Initialize MQTT client
	m.connectToMqttBroker()

	// Main loop to process pending data
	ticker := time.NewTicker(10 * time.Second) // Adjust interval as needed
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down MQTT publisher...")
			return
		case <-ticker.C:
			// Fetch pending data and publish
			m.processPendingData()
		}
	}
}

func (m *MqttPublisher) processPendingData() {
	var pendingData []models.Data

	// Fetch pending data from the database
	if err := m.db.Where("sync_status = ?", "pending").Find(&pendingData).Error; err != nil {
		log.Printf("Error fetching pending data: %v", err)
		return
	}

	for _, data := range pendingData {
		if err := m.publishData(data); err != nil {
			log.Printf("Failed to publish data (ID: %d): %v", data.ID, err)
			continue
		}

		// Update sync_status to "success" after successful publication
		if err := m.db.Model(&data).Update("sync_status", "success").Error; err != nil {
			log.Printf("Failed to update sync_status for data (ID: %d): %v", data.ID, err)
		}
	}
}

// Connect to the MQTT broker with validation and credentials.
func (m *MqttPublisher) connectToMqttBroker() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.brokerURL)
	opts.SetUsername(m.username)
	opts.SetPassword(m.password)
	opts.SetClientID(fmt.Sprintf("client-%d", time.Now().UnixNano()))
	opts.SetCleanSession(true)

	m.client = mqtt.NewClient(opts)

	for {
		token := m.client.Connect()
		if token.Wait() && token.Error() == nil {
			log.Println("Connected to MQTT broker")
			return
		}
		log.Printf("Failed to connect to broker: %v. Retrying in %v...", token.Error(), m.retryDelay)
		time.Sleep(m.retryDelay)
	}
}

// Publish the metric data to the MQTT broker.
func (m *MqttPublisher) publishData(data models.Data) error {
	if m.client == nil || !m.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal data (ID: %d): %v", data.ID, err)
		return err
	}

	token := m.client.Publish(m.topic, 2, false, payload) // QoS = 2
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("failed to publish message (ID: %d): %v", data.ID, token.Error())
	}

	log.Printf("Successfully published data (ID: %d) to MQTT", data.ID)
	return nil
}
