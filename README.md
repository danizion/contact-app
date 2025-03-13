# Rise - Contact Management API

## Overview

A contact management API that allows users to create accounts, authenticate, and manage their personal contacts. The application provides comprehensive CRUD operations for contacts with features like pagination and filtering.

## Architecture
![img.png](img.png)
### Technology Stack

- **Backend**: Go with Gin web framework can be scaled horizontally
- **Database**: PostgreSQL for persistent storage
- **Caching**: Redis for improved performance
- **Authentication**: JWT (JSON Web Tokens)
- **NGINX**: Nginx serving as a reverse proxy and load balancing for app scaling ability

I followed repository pattern which divides the application in to:
1. **Handlers** - The backend entry point to handle HTTP requests and responses
2. **Service Layer**: Contains business logic for user and contact management1.
3. **Repository Layer**: Manages data persistence and retrieval

### How To RUN

1. clone repo:
   ```
   git clone https://github.com/danizion/contact-app.git
   ```
2. In the root directory (where docker-compose.yml is) run:
   ```
   docker-compose -p contacts-app up
   ```
   Congrats! the app is up and running. You should be able to see the app images, redis, docker and nginx images

## Assigment explanation
## On a personal note
I chose to dive into deep waters and develop the app in golang (which made me sweat a fair bit giving this is my first golang experience) and decided to add users to the app to closer imitate a real live application
## How my application meets product definitions:
1. **each contact has**:
  - first name
  - last name
  - phone number
  - address
2. **Services APIs required in my application**:
  - Get contacts --> GET /contacts
  - Search contact --> GET /contacts with wanted query params
  - Add contact --> POST /contacts
  - Edit contact --> PATCH /contacts<contact id>
  - Delete contact --> DELETE /contacts/<contact id>

I decided to adhere to RESTful API Standards (that the reason i dont have an endpoint named /search i.e)

## Bonuses
- Implemented the service in go 
- Added users - each user has its own contacts and every end point access will be handled for the requesting user using the information stored in the JWT token
- Implemented authentication
- Implemented caching mechanism
- designed the system to be able to scale using ngnix to loadblance between server replicas



## API Endpoints

### Authentication

#### User Registration
- **Endpoint**: `POST /users`
- **Description**: Creates a new user account
- **Authentication**: None
- **Request Body**:
  ```json
  {
    "user_name": "johndoe",
    "email": "john@example.com",
    "password": "securepassword"
  }
  ```
- **Response (201 Created)**:
  ```json
  {
    "message": "User created successfully",
    "userID": 123
  }
  ```
- **Error Responses**:
  - `400 Bad Request`: Invalid request body
  - `409 Conflict`: Username or email already exists
  - `500 Internal Server Error`: Server error

#### User Login
- **Endpoint**: `POST /login`
- **Description**: Authenticates a user and returns a JWT token
- **Authentication**: None
- **Request Body**:
  ```json
  {
    "email": "john@example.com",
    "password": "securepassword"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": 123
  }
  ```
- **Error Responses**:
  - `400 Bad Request`: Invalid request body
  - `401 Unauthorized`: Invalid credentials
  - `500 Internal Server Error`: Server error

### Contact Management

All contact management endpoints require authentication using the JWT token obtained from the login endpoint. The token must be included in the `Authorization` header as a Bearer token.

#### Create Contact
- **Endpoint**: `POST /contacts`
- **Description**: Creates a new contact for the authenticated user
- **Authentication**: Required (JWT)
- **Request Headers**:
  ```
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  ```
- **Request Body**:
  ```json
  {
    "first_name": "Jane",
    "last_name": "Smith",
    "phone_number": "123-456-7890",
    "address": "123 Main St, Anytown, USA"
  }
  ```
- **Response (201 Created)**:
  ```json
  {
    "message": "Contact created successfully",
    "contact_id": 456
  }
  ```
- **Error Responses**:
  - `400 Bad Request`: Invalid request body
  - `409 Conflict`: Contact already exists
  - `401 Unauthorized`: Invalid or missing authentication
  - `500 Internal Server Error`: Server error

#### Get Contacts
- **Endpoint**: `GET /contacts`
- **Description**: Retrieves contacts for the authenticated user with pagination and optional filtering
- **Authentication**: Required (JWT)
- **Request Headers**:
  ```
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  ```
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `first_name`: Filter by first name (optional)
  - `last_name`: Filter by last name (optional)
  - `phone_number`: Filter by phone number (optional)
  - `address`: Filter by address (optional)
- **Response (200 OK)**:
  ```json
  {
    "items": [
      {
        "id": 456,
        "user_id": 123,
        "first_name": "Jane",
        "last_name": "Smith",
        "phone_number": "123-456-7890",
        "address": "123 Main St, Anytown, USA"
      }
    ],
    "total_count": 1,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
  ```
- **Error Responses**:
  - `400 Bad Request`: Invalid query parameters
  - `401 Unauthorized`: Invalid or missing authentication
  - `500 Internal Server Error`: Server error

#### Update Contact
- **Endpoint**: `PATCH /contacts/<contact_id>`
- **Description**: Updates an existing contact for the authenticated user
- **Authentication**: Required (JWT)
- **Request Headers**:
  ```
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  ```
- **URL Parameters**:
  - `id`: Contact ID to update
- **Request Body** (all fields optional):
  ```json
  {
    "first_name": "Jane",
    "last_name": "Doe",
    "phone_number": "987-654-3210",
    "address": "456 Oak St, Anytown, USA"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "message": "Contact updated successfully"
  }
  ```
- **Error Responses**:
  - `400 Bad Request`: Invalid request body or contact ID
  - `401 Unauthorized`: Invalid or missing authentication
  - `404 Not Found`: Contact not found
  - `500 Internal Server Error`: Server error

#### Delete Contact
- **Endpoint**: `DELETE /contacts/<contact_id>`
- **Description**: Deletes a contact for the authenticated user
- **Authentication**: Required (JWT)
- **Request Headers**:
  ```
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  ```
- **URL Parameters**:
  - `<contact_id>`: Contact ID to delete
- **Response (200 OK)**:
  ```json
  {
    "message": "Contact deleted successfully"
  }
  ```
- **Error Responses**:
  - `400 Bad Request`: Invalid contact ID
  - `401 Unauthorized`: Invalid or missing authentication
  - `404 Not Found`: Contact not found
  - `500 Internal Server Error`: Server error

## Data Models

### User
- `ID`: Unique identifier (integer - output only)
- `Username`: Unique username (string)
- `Email`: Unique email address (string)
- `Password`: Hashed password (string)

### Contact
- `ID`: Unique identifier (integer - output only)
- `UserID`: Reference to the owner user (integer - output only)
- `FirstName`: First name (string)
- `LastName`: Last name (string)
- `PhoneNumber`: Phone number (string)
- `Address`: Address (string)

## Authentication Flow

1. The client sends a registration request to create a new user account
2. The client sends a login request with email and password
3. If the credentials are valid, the server returns a JWT token
4. The client includes this token in the `Authorization` header for all subsequent requests
5. The server validates the token and identifies the user for each request
6. The token expires after a certain period and the client must obtain a new one

## Caching

The application uses Redis to cache contact data:

1. When a user retrieves contacts, the results are cached with the user ID and filters as the key
2. Subsequent requests with the same parameters will be served from the cache if available
3. When a user creates, updates, or deletes a contact, the cache for that user is invalidated


# Testing
To test the app I created a python script that runs pytests that tests the end point on various tests scenarios. All the scenarios are described in the api_test.py
### How to run the api_tests.py:
#### Prerequisites
1. **Python Installation** Python 3.8 or newer
2. **Install Python Dependencies** 
```
   pip install pytest requests
   ```
3. **Run the tests**
```
   python -m pytest api_tests.py -v
   ```

## API Testing Examples

### User Registration
```bash
curl -X POST http://localhost:80/users \
  -H "Content-Type: application/json" \
  -d '{"user_name":"johndoe","email":"john@example.com","password":"securepassword"}'
```

### User Login
```bash
curl -X POST http://localhost:80/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"securepassword"}'
```

### Create Contact
```bash
curl -X POST http://localhost:80/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"first_name":"Jane","last_name":"Smith","phone_number":"123-456-7890","address":"123 Main St"}'
```

### Get Contacts
```bash
curl -X GET "http://localhost:80/contacts?page=1&first_name=Jane" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update Contact
```bash
curl -X PATCH http://localhost:80/contacts/456 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"first_name":"Jane","last_name":"Doe"}'
```

### Delete Contact
```bash
curl -X DELETE http://localhost:80/contacts/456 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```
