## Telegram AI Bot

This repository contains the source code for a Telegram AI bot written in Go. The bot interacts with users in a Telegram group, responding to messages that mention it. The bot utilizes the GPT-3.5 Turbo model from [Gilas.io](https://gilas.io) to generate responses.

### Functionality

The bot performs the following main functions:

1. **Message Rate Limiting**: It limits the number of responses it gives per user per day to avoid spamming. This limit is configurable.
2. **Integration with Redis**: It uses Redis as a data store for managing message rate limiting.
3. **Integration with Telegram API**: It communicates with Telegram's API to send and receive messages.
4. **Integration with Gilas.io API**: It connects to Gilas.io's API to generate AI-driven responses using the GPT-3.5 Turbo model.
5. **Graceful Shutdown**: It gracefully handles signals for termination, ensuring clean shutdown when needed.

### Parameters

The bot requires the following parameters to be configured:

- **Telegram API Key**: The API key for accessing the Telegram Bot API.
- **Telegram Group ID**: The ID of the Telegram group where the bot operates.
- **Gilas.io API Key**: The API key for accessing the Gilas.io API.
- **Redis Configuration**: Address and password for connecting to the Redis database.
- **Message Rate Limit**: The maximum number of responses allowed per user per day.

### Usage

To use this code run:

```
docker-compose run
```

### License

This project is licensed under the [MIT License](LICENSE). Feel free to modify and distribute it as per the terms of the license.

For detailed implementation and usage instructions, refer to the code comments and documentation within the source files.
