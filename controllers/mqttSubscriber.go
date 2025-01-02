package controllers

import (
	"billingo/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MQTTSubscriber handles subscribing to an MQTT topic and saving data to the database.
type MQTTSubscriber struct {
	client mqtt.Client
	db     *gorm.DB
	topic  string
}

// NewMQTTSubscriber creates a new MQTTSubscriber.
func NewMQTTSubscriber(broker, topic string) *MQTTSubscriber {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(fmt.Sprintf("mqtt-subscriber-%d", time.Now().Unix()))
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Infof("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
	})
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Warnf("Connection lost: %v", err)
	}
	opts.OnConnect = func(client mqtt.Client) {
		log.Infof("Connected to broker: %s", broker)
	}
	client := mqtt.NewClient(opts)

	return &MQTTSubscriber{
		client: client,
		topic:  topic,
	}
}

// Setup prepares the subscriber.
func (s *MQTTSubscriber) Setup(db *gorm.DB, manager *Manager) {
	s.db = db
	go func() {
		for {
			if token := s.client.Connect(); token.Wait() && token.Error() != nil {
				log.Errorf("Failed to connect to MQTT broker: %v", token.Error())
				log.Infof("Retrying connection in 10 seconds...")
				time.Sleep(10 * time.Second)
				continue
			}
			log.Infof("MQTTSubscriber connected to broker")
			return // Exit the retry loop after successful connection
		}
	}()
}

// String returns a string representation of the MQTTSubscriber.
func (s *MQTTSubscriber) String() string {
	return fmt.Sprintf("MQTTSubscriber[topic=%s]", s.topic)
}

// Main subscribes to the topic and saves received messages to the database.
func (s *MQTTSubscriber) Main(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("MQTTSubscriber stopping...")
			// Unsubscribe and disconnect gracefully
			if token := s.client.Unsubscribe(s.topic); token.Wait() && token.Error() != nil {
				log.Warnf("Failed to unsubscribe from topic: %v", token.Error())
			}
			s.client.Disconnect(250)
			return
		default:
			// Attempt to subscribe
			log.Infof("Attempting to subscribe to topic: %s", s.topic)
			token := s.client.Subscribe(s.topic, 1, func(client mqtt.Client, msg mqtt.Message) {
				// Unmarshal the message payload into RRDData
				var vmData models.VMData
				if err := json.Unmarshal(msg.Payload(), &vmData); err != nil {
					log.Errorf("Failed to unmarshal MQTT message payload: %v", err)
					return
				}

				// Save to data_raw table
				dataRaw := models.DataRaw{
					RRDData: models.RRDData{
						Time:      vmData.Time,
						MaxCPU:    vmData.MaxCPU,
						MaxDisk:   vmData.MaxDisk,
						MaxMem:    vmData.MaxMem,
						Disk:      vmData.Disk,
						CPU:       vmData.CPU,
						Mem:       vmData.Mem,
						NetOut:    vmData.NetOut,
						NetIn:     vmData.NetIn,
						DiskRead:  vmData.DiskRead,
						DiskWrite: vmData.DiskWrite,
					},
					VMID:  vmData.VMID,
					Topic: s.topic,
				}
				if err := s.db.Create(&dataRaw).Error; err != nil {
					log.Errorf("Failed to save data to database: %v", err)
				} else {
					log.Infof("Data saved to database: %v", dataRaw)
				}
			})

			// Check if subscription was successful
			if token.Wait() && token.Error() != nil {
				log.Errorf("Failed to subscribe to topic: %v", token.Error())
				log.Infof("Retrying in 10 seconds...")
				time.Sleep(10 * time.Second) // Retry after 10 seconds
				continue
			}

			log.Infof("Successfully subscribed to topic: %s", s.topic)

			// Block until the context is canceled
			<-ctx.Done()
			log.Infof("MQTTSubscriber stopping...")
			if token := s.client.Unsubscribe(s.topic); token.Wait() && token.Error() != nil {
				log.Warnf("Failed to unsubscribe from topic: %v", token.Error())
			}
			s.client.Disconnect(250)
			return
		}
	}
}
