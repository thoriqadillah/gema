# Example Project
The following are the steps to setup the project
1. Setup environment variables
Make sure to create `.env` file for easier environment variable

2. Migrate the migrations
On another terminal, run the following
```bash
cd cmd
go run . migrate up
```

3. Run the app
```bash
go run .
```