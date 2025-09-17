// Package internal provides functions for importing data
package internal

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	min "worker-service/pkg"
	log "worker-service/utils"

	"github.com/xuri/excelize/v2"
)

type Product struct {
	ID    int
	Name  string
	Price float64
}

func ImportData(key string, ctx context.Context, db *sql.DB, minioClient *min.MinioClient) {
	fmt.Printf("📦 Job key: %s\n", key)

	// Get the object from Minio
	obj := minioClient.GetObject(key)
	defer obj.Close()

	// Open the file
	file, err := excelize.OpenReader(obj)
	log.FailOnError(err, "Failed to open file")
	defer file.Close()

	// Get the rows
	sheetName := "Sheet1"
	rows, err := file.GetRows(sheetName)
	log.FailOnError(err, "Failed to get rows")

	newRows := rows[:]

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	txOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}
	tx, err := db.BeginTx(ctxTimeout, txOpts)
	log.FailOnError(err, "Failed to begin transaction")

	defer func() {
		if err := recover(); err != nil {
			err := tx.Rollback()
			log.FailOnError(err, "Failed to rollback transaction")
			fmt.Println("❌ Rolled back transaction")
		} else {
			err := tx.Commit()
			log.FailOnError(err, "Failed to commit transaction")
			fmt.Println("✅ Committed transaction")
		}
	}()

	for index, row := range newRows {
		product := Product{Name: row[1]}

		if _, err := fmt.Sscan(row[0], &product.ID); err != nil {
			fmt.Printf("Row %d: error parsing id: %v", index+1, err)
			continue
		}

		if _, err := fmt.Sscan(row[2], &product.Price); err != nil {
			fmt.Printf("Row %d: error parsing price: %v", index+1, err)
			continue
		}

		const rawQuery = `
			INSERT INTO product(id, name, price)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				price = VALUES(price)
		`
		result, err := tx.ExecContext(
			ctxTimeout,
			rawQuery,
			product.ID, product.Name, product.Price,
		)
		log.FailOnError(err, "Failed to execute query")

		lastInsertID, err := result.LastInsertId()
		log.FailOnError(err, "Failed to get last insert id")
		effectedRows, err := result.RowsAffected()
		log.FailOnError(err, "Failed to get rows affected")

		if lastInsertID != 0 {
			fmt.Printf("✅ Inserted row with id: %d\n", lastInsertID)
		}

		if effectedRows != 1 {
			fmt.Printf("❌ Inserted %d rows\n", effectedRows)
		}
	}

	fmt.Println("Length of rows: ", len(rows))
}
