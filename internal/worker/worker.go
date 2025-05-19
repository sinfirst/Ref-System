package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sinfirst/Ref-System/internal/models"
	"github.com/sinfirst/Ref-System/internal/storage/pg"
)

func PollOrderStatus(ctx context.Context, orderNum, user string, accrual string, storage *pg.PGDB) {
	url := fmt.Sprintf("%s/api/orders/%s", accrual, orderNum)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(2 * time.Minute)
	attempts := 0
	maxAttempts := 12

	for {
		select {
		case <-ticker.C:
			var responce models.OrderResponce
			attempts++

			if attempts > maxAttempts {
				fmt.Println("Превышено количество попыток")
				return
			}
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("poll error:", err)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fmt.Println("read response error:", err)
				continue
			}

			err = json.Unmarshal(body, &responce)
			if err != nil {
				fmt.Println("Ошибка парсинга:", err)
				return
			}

			if responce.Status == "PROCESSED" {
				err := storage.UpdateStatus(ctx, responce.Status, responce.Order, user)

				if err != nil {
					fmt.Println("Error in update db: ", err)
					return
				}

				err = storage.UpdateUserBalance(ctx, user, float32(responce.Accrual), 0)

				if err != nil {
					fmt.Println("Error in update db: ", err)
					return
				}
				return
			} else {
				fmt.Println(responce.Status, " not equal PROCESSED!")
			}
		case <-timeout:
			fmt.Println("Time is out")
			return
		}
	}
}
