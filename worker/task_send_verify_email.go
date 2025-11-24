package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/util"
	"github.com/rs/zerolog/log"
)

// all tasks will be stored

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

// will add tasks to the queue
func (distributor *RedisTaskDistributor) DistributeTaskSendBerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// created task
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	// enqueued task
	taskInfo, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", taskInfo.Queue).Int("max_retry", taskInfo.MaxRetry).
		Msg("enqueued task")

	return nil
}

// will take tasks from the queue and process them
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload PayloadSendVerifyEmail

	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	// get user from database
	user, err := processor.store.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// create email record in DB
	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(10),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	// Send Email to user
	subject := "Welcome to Message App"
	verifyUrl := fmt.Sprintf(
		"http://localhost:8080/verify_email?email_id=%d&secret_code=%s", verifyEmail.EmailID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello %s, <br/>
	Thank you for registering with us! <br/>
	Please <a href="%s"> click here</a> to verify yor email address <br/>
	`, user.Username, verifyUrl)
	to := []string{user.Email}

	processor.mailer.SendEmail(subject, content, to, nil, nil, nil)

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")

	return nil
}
