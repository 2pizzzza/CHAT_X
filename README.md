
# Chat Application

This is a chat application written in Golang using Fiber, GORM, and JWT authentication.

## Endpoints

### Authentication

- `POST /auth/register` - Register a new user.
- `POST /auth/login` - Log in a user.
- `GET /auth/logout` - Log out the current user.
- `POST /auth/confirm` - Confirm user registration.
- `POST /auth/refresh` - Refresh JWT access token.
- `POST /auth/change-password` - Change user password.
- `GET /auth/verify-email` - Verify user email.
- `POST /auth/reset-password-request` - Request to reset user password.
- `POST /auth/reset-password-verify` - Verify reset password request.
- `POST /auth/reset-password` - Reset user password.

### Chat

- `POST /chat/messages` - Create a new message.
- `POST /chat/delete-messages` - Delete message(s).
- `GET /chat/ws/:chatID` - WebSocket endpoint for chat communication.
- `PUT /chat/change-messages` - Update message(s).

### User

- `GET /users/me` - Get current user information.

### Health Checker

- `GET /healthchecker` - Check the health of the application.

## WebSocket

The application uses WebSocket for real-time chat communication. WebSocket connections are established at `/chat/ws/:chatID`.

## Usage

To use this application, make sure you have Go installed. Clone the repository and run the following commands:

```sh
go mod download
go run main.go
