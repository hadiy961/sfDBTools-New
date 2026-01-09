package helpers

import (
	"context"
	"fmt"
	"sync"
	"time"

	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/database"

	"github.com/dustin/go-humanize"
)

func collectSingleDatabaseDetail(ctx context.Context, client *database.Client, logger applog.Logger, dbName string, sizeProvider func(context.Context, string) (int64, error)) dbscanmodel.DatabaseDetailInfo {
	startTime := time.Now()
	detail := dbscanmodel.DatabaseDetailInfo{
		DatabaseName:   dbName,
		CollectionTime: startTime.Format("2006-01-02 15:04:05"),
	}

	type metricResult struct {
		metricType string
		value      int64
		err        error
	}

	metricChan := make(chan metricResult, 6)
	var metricWg sync.WaitGroup

	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		var (
			size int64
			err  error
		)
		if sizeProvider != nil {
			size, err = sizeProvider(ctx, dbName)
		} else {
			size, err = client.GetDatabaseSize(ctx, dbName)
		}
		metricChan <- metricResult{"size", size, err}
	}()

	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		count, err := client.GetTableCount(ctx, dbName)
		metricChan <- metricResult{"tables", int64(count), err}
	}()

	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		count, err := client.GetProcedureCount(ctx, dbName)
		metricChan <- metricResult{"procedures", int64(count), err}
	}()

	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		count, err := client.GetFunctionCount(ctx, dbName)
		metricChan <- metricResult{"functions", int64(count), err}
	}()

	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		count, err := client.GetViewCount(ctx, dbName)
		metricChan <- metricResult{"views", int64(count), err}
	}()

	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		count, err := client.GetUserGrantCount(ctx, dbName)
		metricChan <- metricResult{"user_grants", int64(count), err}
	}()

	go func() {
		metricWg.Wait()
		close(metricChan)
	}()

	var errors []string
	for result := range metricChan {
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", result.metricType, result.err))
			continue
		}

		switch result.metricType {
		case "size":
			detail.SizeBytes = result.value
			detail.SizeHuman = humanize.Bytes(uint64(result.value))
		case "tables":
			detail.TableCount = int(result.value)
		case "procedures":
			detail.ProcedureCount = int(result.value)
		case "functions":
			detail.FunctionCount = int(result.value)
		case "views":
			detail.ViewCount = int(result.value)
		case "user_grants":
			detail.UserGrantCount = int(result.value)
		}
	}

	if len(errors) > 0 {
		detail.Error = fmt.Sprintf("Errors: %v", errors)
		logger.Warningf("Error mengumpulkan detail database %s: %v", dbName, errors)
	}

	return detail
}
