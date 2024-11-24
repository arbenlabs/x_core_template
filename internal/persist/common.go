package persist

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

func InsertRecord[T any](db *gorm.DB, record T) (*T, error) {
	result := db.Create(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func BatchInsert[T any](db *gorm.DB, records []T, batchSize int) error {
	if err := db.CreateInBatches(records, batchSize).Error; err != nil {
		return err
	}
	return nil
}

func GetAllRecords[T any](db *gorm.DB, page, pageSize int) ([]T, int, error) {
	var records []T
	var totalRecords int64

	// Count total number of records in the table
	if err := db.Model(new(T)).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := db.Offset(offset).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Calculate the total number of pages
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))

	return records, totalPages, nil
}

func GetRecordByID[T any](db *gorm.DB, id string) (*T, error) {
	var record T
	result := db.Where("id = ?", id).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func GetRecordByField[T any](db *gorm.DB, fieldName string, fieldValue interface{}) (*T, error) {
	var record T

	result := db.Where(fmt.Sprintf("%s = ?", fieldName), fieldValue).First(&record)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil if no record is found
		}
		return nil, result.Error // Return the error for other cases
	}

	return &record, nil
}

func GetRecordsByField[T any](db *gorm.DB, field string, value interface{}, page, pageSize int, orderBy string) ([]T, int64, error) {
	var records []T
	var totalCount int64

	// Count total records
	countQuery := db.Model(new(T)).Where(fmt.Sprintf("%s = ?", field), value)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Prepare query
	query := db.Where(fmt.Sprintf("%s = ?", field), value)
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// Execute query
	result := query.Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return records, totalCount, nil
}

func GetRecordsByFields[T any](db *gorm.DB, conditions map[string]interface{}) ([]T, error) {
	var records []T

	query := db
	for field, value := range conditions {
		query = query.Where(field+" = ?", value)
	}

	result := query.Find(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

func GetFilteredPaginatedRecords[T any](db *gorm.DB, page, pageSize int, conditions map[string]interface{}) ([]T, int, error) {
	var records []T
	var totalRecords int64

	query := db.Model(new(T)) // Apply model to the query for proper counting

	// Apply conditions dynamically
	for field, value := range conditions {
		if field == "price" {
			str, ok := value.(string)
			if !ok {
				return nil, 0, fmt.Errorf("value (price) is not a string")
			}

			// Handle "+" and "-" at the end of the string
			lastChar := str[len(str)-1]
			priceStr := str[:len(str)-1] // Remove last character for conversion

			price, err := strconv.Atoi(priceStr)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid price value: %v", err)
			}

			if string(lastChar) == "-" {
				query = query.Where(field+" <= ?", price)
				continue
			}

			if string(lastChar) == "+" {
				query = query.Where(field+" >= ?", price)
				continue
			}
		}

		if field == "pct_remaining" {
			str, ok := value.(string)
			if !ok {
				return nil, 0, fmt.Errorf("value (pct_remaining) is not a string")
			}

			// Handle "+" and "-" at the end of the string
			lastChar := str[len(str)-1]
			pctStr := str[:len(str)-1] // Remove last character for conversion

			pctRemaining, err := strconv.ParseFloat(pctStr, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid pct_remaining value: %v", err)
			}

			if string(lastChar) == "-" {
				query = query.Where(field+" <= ?", pctRemaining)
				continue
			}

			if string(lastChar) == "+" {
				query = query.Where(field+" >= ?", pctRemaining)
				continue
			}
		}

		// Default condition for equality
		query = query.Where(field+" = ?", value)
	}

	// Count total number of records after applying conditions
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Calculate total pages
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return records, totalPages, nil
}

func UpdateRecordByID[T any, U any](db *gorm.DB, id string, updates U) error {
	var record T
	result := db.Model(&record).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteRecordByID[T any](db *gorm.DB, id string) error {
	var record T
	result := db.Where("id = ?", id).Delete(&record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
