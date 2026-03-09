# Gema
Gema is a simplified version of [Echo](https://echo.labstack.com/) http framework by using dependency injenction using [fx](https://github.com/uber-go/fx). If you are familiar with Nest.js, it's kinda the same. What's included:
- Logger module with zap logger
- Command line module to create CLI app with [cobra](https://github.com/spf13/cobra)
- Database module using [bun](https://bun.uptrace.dev/) with [pgxpool](https://github.com/jackc/pgx/tree/master/pgxpool) as the connection pool
- Seeding command with the capability to register your seeder
- Migration command with [goose](https://github.com/pressly/goose)
- Storage module. Currently only local storage using your file system. Suitable for local development. But you can register your own storage like S3, Google Cloud Storage, etc
- Notifier module, like email notification. Currently only email notification is available
- Easier validation with `gema.Validator` interface and `gema.Validate` to validate your struct after struct binding
- Easier to create a controller with `gema.Controller` interface
- GRPC server
- Message queue module using river queue

## Usage
Please see example folder for how to use any of the available utilities

## Next
- [ ] Auth
- [ ] RBAC
- [ ] Websocket module
- [ ] Obersvability module
