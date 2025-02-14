<h1 align="center">Peril</h1>

<div align="center">

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

</div>

## Description

Peril is an interactive CLI game developed in Go that leverages a Pub/Sub architecture powered by RabbitMQ. The project includes a server capable of pausing and resuming the game, and multiple clients that allow players to interact using simple text commands. Designed as both a fun challenge and a learning tool, Peril offers insights into real-time messaging systems and distributed architectures.

## Motivation

The primary motivation behind Peril is to explore and gain practical experience with Pub/Sub architectures. By integrating RabbitMQ and Go, the project provides a hands-on opportunity to understand how message queues can be used to build responsive, decoupled systems. Whether you're a developer curious about distributed systems or simply looking to create an engaging CLI game, Peril offers an ideal playground for experimentation and learning.

## Quick Start

### Prerequisites

- **Go:** Ensure you have Go installed (version 1.23 or higher is recommended). You can download it from the [official Go website](https://golang.org/dl/).
- **Docker:** Required to build the RabbitMQ image.
- **RabbitMQ:** The game relies on RabbitMQ running as a container.

### Build and Run RabbitMQ

1. **Build the Docker Image:**

   Navigate to the project root (where the `Dockerfile` is located) and build the RabbitMQ image:

   ```bash
   docker build -t rabbitmq .
   ```

2. **Start RabbitMQ:**

   Run the provided script to start RabbitMQ:

   ```bash
   ./rabbit.sh
   ```

   This script will launch the RabbitMQ container necessary for the game's messaging system.

### Build and Run the Server

1. **Compile the Server:**

   Navigate to the `cmd/server` directory and build the server binary:

   ```bash
   cd server
   go build -o server .
   ```

2. **Run the Server:**

   Start the server by executing:

   ```bash
   ./server
   ```

   The server will initialize the Pub/Sub channels and is capable of pausing/resuming the game as required.

### Build and Run the Client

1. **Compile the Client:**

   Navigate to the `cmd/client` directory and build the client binary:

   ```bash
   cd client
   go build -o client .
   ```

2. **Run the Client:**

   Launch the client by executing:

   ```bash
   ./client
   ```

   When the client starts, you'll be prompted to enter your username, after which you can start issuing commands.

### Note

- **Environment Variables:** This project does not rely on any `.env` files. All configurations are handled within the source code or via command-line parameters.
- **Docker & RabbitMQ:** Ensure Docker is running on your machine before executing the `rabbit.sh` script.

## Detailed Features

### Game Server

- **Game Control:** The server manages the overall state of the game and can pause/resume gameplay.
- **Pub/Sub Integration:** Uses RabbitMQ to handle message queuing between the server and clients.

### Client Commands

Upon launching, the client requires a username and supports the following commands:

- **spawn:** Spawn a new unit at a specific location.
- **move:** Move a spawned unit to a specific location.
- **status:** Display the current status and statistics of the player.
- **help:** Print a help message outlining available commands and usage.
- **spam:** Send a flood of messages into the queue for fun (or mischief).
- **quit:** Exit the game.

Each command is processed and communicated to the server via RabbitMQ, ensuring a decoupled and responsive gaming experience.

## 🤝 Contributing

Contributions are welcome! If you'd like to contribute to Peril, please fork the repository and open a pull request against the `main` branch. For major changes, please open an issue first to discuss what you would like to change.

Happy coding and enjoy the game!
