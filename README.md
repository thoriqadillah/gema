# Gema
Gema is a simplified version of [Echo](https://echo.labstack.com/) http framework by using dependency injenction using [fx](https://github.com/uber-go/fx). If you are familiar with Nest.js, it's kinda the same. What's included:
- Logger module with zap logger
- Command line module to create CLI app with [cobra](https://github.com/spf13/cobra)
- Database module using [bun](https://bun.uptrace.dev/) with [pgxpool](https://github.com/jackc/pgx/tree/master/pgxpool) as the connection pool
- Transactional CLS module to propagate tx instance in request scope. It's almost the same with [Nest.js transactional CLS](https://papooch.github.io/nestjs-cls/plugins/available-plugins/transactional) module if you are familiar with it
- Seeding command with the capability to register your seeder
- Migration command with [goose](https://github.com/pressly/goose)
- Storage module. Currently only local storage using your file system. Suitable for local development. But you register your own storage like S3, Google Cloud Storage, etc
- Message queue using [river queue](https://riverqueue.com/)
- Notifier module, like email notification. Currently only email notification is available. If you are using the river queue module, email sending with `RiveredEmailNotifier` name can be use to send the email using river queue. But you can register your own notifier
- Easier validation with `gema.Validator` interface to validate your struct after struct binding
- Easier to create a controller with `gema.Controller` interface

## Usage
Please see example folder for how to use any of the available utilities

## Next
- [ ] Cache module
- [ ] gRPC module
- [ ] Websocket module
- [ ] Obersvability module