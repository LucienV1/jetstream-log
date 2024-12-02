# Jetstream Log

This program subscribes to the atproto jetstream, and logs all events by default. this could be used for a variety of purposes, like training an llm.

## Features

- Connects to a Jetstream WebSocket server
- Subscribes to custom event lists and logs them into a database
- Supports SQLite and PostgreSQL databases
- Filters messages based on specified parameters

## Requirements

- Go 1.23 (it might work on earlier versions, but I've only tested it with 1.23)
- SQLite or PostgreSQL database

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/LucienV1/jetstream-log.git
    cd jetstream-log
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Build the application:

    ```sh
    go build -o jetstream-log
    ```

## Usage

Run the application with the following command:

```sh
./jetstream-log -t <database_type> -s <sqlite_db_path> -p <postgres_connection_string> -q <filter_params> -w <wanted_dids> -r <wss_uri>
```

### Command Line Arguments

- `-t`: Database type (`sqlite` or `postgres`). Default is `sqlite`.
- `-s`: SQLite database path. Default is `output.sqlite`.
- `-p`: PostgreSQL connection string. Default is `postgres://postgres:password@localhost:5432/postgres`.
- `-q`: Array of strings to filter the messages, by collection. Default is `["app.bsky.*"]`. Be sure to use the `["example.example", "example.example2.example"]` format.
- `-w`: Array of DIDs to filter the messages. No default, and uses the same format as the above option.
- `-r`: WebSocket URI to connect to. Default is `wss://jetstream1.us-east.bsky.network/subscribe`.

### Examples

```sh
./jetstream-log -t sqlite -s output.sqlite -q '["app.bsky.*"]' -r wss://jetstream1.us-east.bsky.network/subscribe
```

to use defaults, but only log posts, you can use 
```sh
./jetstream-log -q '["app.bsky.feed.post"]'
```

## Data Storage

By default, the application gathers approximately 7GB of data per day. Ensure you have sufficient storage space and manage your database size accordingly.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under AGPLv3. See the [LICENSE](LICENSE) file for details.
