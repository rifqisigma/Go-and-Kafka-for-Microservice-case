package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"service_notification/dto"
	"service_notification/helper/utils"
	"time"

	"github.com/segmentio/kafka-go"
)

func ProductResponseConsumer() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "notification-request",
		GroupID: "notification-service",
	})

	go func() {
		for {
			msg, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Println("error reading message:", err)
				continue
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				fmt.Println("error unmarshaling:", err)
				continue
			}

			corrID := payload["correlation_id"].(string)
			service := payload["service"].(string)
			action := payload["action"].(string)
			email := payload["email"].(string)
			message := payload["message"]

			fmt.Println(message)

			if service == "user" {
				if action == "register" {
					html := fmt.Sprintf("<h1>ActionId:%s <br> Hello %s selamat datang di shop!</h1>", corrID, message.(string))
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)
				}
			} else if service == "store" {
				if action == "create" {
					var store dto.Store
					err := json.Unmarshal([]byte(message.(string)), &store)
					if err != nil {
						fmt.Println(err)
					}
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil membuat store <br> id:%d <br> name:%s <br> admin created:%d <br>created by :%s</h1>", corrID, store.ID, store.Name, store.AdminID, store.CreatedAt.Format(time.RFC1123))
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)

				} else if action == "update" {
					var store dto.Store
					err := json.Unmarshal([]byte(message.(string)), &store)
					if err != nil {
						fmt.Println(err)
					}
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil mengupdate store <br> id:%d <br> name:%s </h1>", corrID, store.ID, store.Name)
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)
				} else if action == "delete" {
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil menghapus store dengan id %v </h1>", corrID, message.(uint))
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)
				}
			} else if service == "product" {
				if action == "create" {
					var product dto.Product
					err := json.Unmarshal([]byte(message.(string)), &product)
					if err != nil {
						fmt.Println(err)
					}
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil menambahkan product <br> id:%d <br> name:%s <br> stock:%d <br>created by :%s</h1>", corrID, product.ID, product.Name, product.Stock, product.CreatedAt.Format(time.RFC1123))
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)

				} else if action == "update" {
					var product dto.Product
					err := json.Unmarshal([]byte(message.(string)), &product)
					if err != nil {
						fmt.Println(err)
					}
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil mengupdate product <br> id:%d <br> name:%s <br> stock:%d</h1>", corrID, product.ID, product.Name, product.Stock)
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)
				} else if action == "delete" {
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil menghapus product dengan id %v </h1>", corrID, message.(uint))
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)
				}
			} else if service == "cart" {
				if action == "paid" {
					var paid dto.UpdatePaidCartItemReq
					err := json.Unmarshal([]byte(message.(string)), &paid)
					if err != nil {
						fmt.Println(err)
					}
					html := fmt.Sprintf("<h1>ActionId:%s <br>anda berhasil membeli product <br> cart id:%d <br> product_id:%d <br> total:%d <br>buy date :%s</h1>", corrID, paid.ID, paid.ProductID, paid.PurchaseAmount, time.Now().Format(time.RFC1123))
					send := dto.SendEmail{
						ToEmail:  email,
						Header:   service,
						ActionId: corrID,
						Desc:     html,
					}
					utils.SendEmail(&send)
				}
			}
		}
	}()
}
