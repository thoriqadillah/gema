package gema

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

const RiveredEmailNotifier NotifierName = "riveredemail"

type RiveredEmailer struct {
	river *river.Client[pgx.Tx]
	pool  *pgxpool.Pool
}

type emailArg struct {
	Message
}

func (emailArg) Kind() string {
	return "email"
}

func createRiveredEmailer(o *NotifierOption) Notifier {
	return &RiveredEmailer{
		river: o.River,
		pool:  o.Pool,
	}
}

func (e *RiveredEmailer) Send(ctx context.Context, m Message) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}

	_, err = e.river.InsertTx(ctx, tx, emailArg{m}, &river.InsertOpts{
		MaxAttempts: 3,
		Priority:    river.PriorityDefault,
		Queue:       "notification",
	})

	if err != nil {
		return tx.Rollback(ctx)
	}

	return tx.Commit(ctx)
}

type emailWorker struct {
	emailer Notifier
	river.WorkerDefaults[emailArg]
}

func (w *emailWorker) Work(ctx context.Context, job *river.Job[emailArg]) error {
	return w.emailer.Send(ctx, job.Args.Message)
}

func init() {
	RegisterNotifier(RiveredEmailNotifier, createRiveredEmailer)
}
