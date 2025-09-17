// Package jobs provider jobs for workers
package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"worker-service/internal"
	min "worker-service/pkg"
)

type MessagePayload struct {
	Key    string `json:"key"`
	Result any    `json:"result"`
}

type Job struct {
	ID          int
	Payload     MessagePayload
	DB          *sql.DB
	MinioClient *min.MinioClient
}

func (j *Job) Process(ctx context.Context) {
	fmt.Printf("🏁 Job %d started\n", j.ID)

	select {
	case <-ctx.Done():
		fmt.Printf("❌ Job %d cancelled\n", j.ID)
		return
	default:
		key := j.Payload.Key
		internal.ImportData(key, ctx, j.DB, j.MinioClient)

		fmt.Printf("✅ Job %d done\n", j.ID)
	}
}
