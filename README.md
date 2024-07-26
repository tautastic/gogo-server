# GoGo-Server

## Features

- **Check Server Readiness**: Verify if the server is ready to process requests.
- **Initialize the Game**: Set up the game board, komi, and handicaps.
- **Generate Moves**: Request AI-generated moves from KataGo.
- **Play Moves**: Submit player moves to the server.

## Installation

1. **Setup KataGo**:
    Refer to the [KataGo repository](https://github.com/lightvector/KataGo) for installation instructions.

2. **Clone the repository**:
    ```sh
    git clone https://github.com/tautastic/gogo-server.git
    cd gogo-server
    ```

3. **Set up environment variables**:
    - `AUTHORIZATION_TOKEN`: Token for authenticating API requests.
    - `KATAGO_PATH`: Path to the KataGo executable.

4. **Run the server**:
    ```sh
    go run main.go
    ```

## API Endpoints

### `GET /check-ready`

Check if the server is ready to process requests.

- **Response**: `200 OK` if ready, `503 Service Unavailable` if not ready.

### `POST /init`

Initialize the game with the specified parameters.

- **Request Body**:
    ```json
    {
        "board-size": "13",
        "komi": "7.5",
        "handicaps": ["D4", "F11"]
    }
    ```

### `POST /play-move`

Submit a move to the server.

- **Request Body**:
    ```json
    {
        "color": "white",
        "move_to_pos": "D4"
    }
    ```

### `POST /gen-move`

Request an AI-generated move.

- **Request Body**:
    ```json
    {
        "color": "black"
    }
    ```

- **Response Body**:
    ```json
    {
        "move": "D4"
    }
    ```

## Project Structure

- **main.go**: Entry point of the application. Loads environment variables, starts the KataGo process, and initializes the API server.
- **api/api.go**: Contains the HTTP handlers for the API endpoints.
- **cmd/cmd.go**: Manages the interaction with the KataGo process, including reading and writing to its stdin and stdout.

## Purpose

This project was developed as part of an effort to automate the IPvGo minigame in the [Bitburner](https://github.com/bitburner-official/bitburner-src) video game.
See the `ipvgo.js` file for one possible usage of the gogo-server in Bitburner.
