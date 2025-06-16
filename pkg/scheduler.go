package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"log"
	"time"
)

func mustJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func CreateReminderSchedules(
	ctx context.Context,
	schedulerClient *scheduler.Client,
	lambdaArn string,
	roleArn string,
	reminderID string,
	reminderTime time.Time,
	payload map[string]string,
) error {
	times := []struct {
		offset time.Duration
		suffix string
	}{
		{24 * time.Hour, "24h"},
		{1 * time.Hour, "1h"},
		{10 * time.Minute, "10m"},
	}

	for _, t := range times {
		scheduleTime := reminderTime.Add(-t.offset)
		log.Printf("Creating schedule for reminder %s at %s (offset: %s)", reminderID, scheduleTime, t.suffix)
		if scheduleTime.Before(time.Now()) {
			continue // не создавать напоминание в прошлом
		}
		// Имя расписания: userID_reminderID_unixtime
		userID := payload["user_id"]
		unixTime := fmt.Sprintf("%d", scheduleTime.Unix())
		scheduleName := userID + "_" + reminderID + "_" + unixTime
		input := &scheduler.CreateScheduleInput{
			Name:               &scheduleName,
			ScheduleExpression: aws.String("at(" + scheduleTime.UTC().Format("2006-01-02T15:04:05") + ")"),
			FlexibleTimeWindow: &types.FlexibleTimeWindow{
				Mode:                   types.FlexibleTimeWindowModeFlexible,
				MaximumWindowInMinutes: aws.Int32(2),
			},
			Target: &types.Target{
				Arn:     &lambdaArn,
				RoleArn: &roleArn,
				Input:   aws.String(mustJson(payload)),
			},
		}
		_, err := schedulerClient.CreateSchedule(ctx, input)
		if err != nil {
			return err
		}
	}
	return nil
}
