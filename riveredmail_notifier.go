package gema

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
)

const RiveredEmailNotifier NotifierName = "riveredemail"

type riveredEmailer struct {
	river *river.Client[pgx.Tx]
	pool  *pgxpool.Pool
}

type emailArg struct {
	Message
}

func (emailArg) Kind() string {
	return "email"
}

func createRiveredEmailer(river *river.Client[pgx.Tx], pool *pgxpool.Pool) Notifier {
	return &riveredEmailer{
		river: river,
		pool:  pool,
	}
}

func (e *riveredEmailer) Send(ctx context.Context, m Message) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}

	_, err = e.river.InsertTx(ctx, tx, emailArg{m}, &river.InsertOpts{
		MaxAttempts: 3,
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

type riveredEmailProvider struct {
	opt *EmailerOption
}

func RiveredEmailProvider(opt *EmailerOption) NotifierProvider {
	return &riveredEmailProvider{
		opt: opt,
	}
}

func (p *riveredEmailProvider) registerProvider(notifier Notifier, registry NotifierRegistry) {
	registry.Register(RiveredEmailNotifier, notifier)
}

func (p *riveredEmailProvider) Register() fx.Option {
	return fx.Module("notifier.rivered_emailer",
		fx.Provide(fx.Private, createRiveredEmailer),
		fx.Invoke(p.registerProvider),
		fx.Invoke(func() {
			river.AddWorker(workers, &emailWorker{
				emailer: newEmailNotifier(p.opt),
			})
		}),
	)
}
