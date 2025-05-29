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
			var response models.OrderResponce
			attempts++

			if attempts > maxAttempts {
				fmt.Println("Превышено количество попыток")
				return
			}
			resp, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
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

			err = json.Unmarshal(body, &response)
			if err != nil {
				fmt.Println("Ошибка парсинга:", err)
				return
			}

			if response.Status == "PROCESSED" {
				err := storage.UpdateStatus(ctx, response.Status, response.Order, user)

				if err != nil {
					fmt.Println("Error in update db: ", err)
					return
				}

				err = storage.UpdateUserBalance(ctx, user, float32(response.Accrual), 0)

				if err != nil {
					fmt.Println("Error in update db: ", err)
					return
				}
				return
			} else {
				fmt.Println(response.Status, " not equal PROCESSED!")
			}
		case <-timeout:
			fmt.Println("Time is out")
			return
		case <-ctx.Done():
			fmt.Println("Done")
			return
		}
	}
}
