## Setup

To configure the server, rename the .env_dist to .env and edit the variables.

To create the SQLite database, use the `npm i` script followed by `npm run setup`. This will create the database. To further customise the database entries, you can look at `setupDatabase.js`.

## Authentication

To authenticate yourself, please use the `/api/auth/login` route. It uses basic auth and will create a JWT token. All other routes are only accessible with this token.

## Environment variables

Firstly, the environment variables given into the container are used. Afterwards, the variables from the .env file are used. Currently, there are four variables to set

```
API_ADMIN=admin
API_PASSWORD=admin
```
To set the primary admin.

DEMO_DATA=FALSE
This creates demo data to play around with the backend. The demo data creates three washing machines and a display user, which is useful for testing the frontend.

```PORT=3000```
Set the desired port

## Creating users

To create users, use the `/api/users` route.

To create a user, a type is needed.
Currently, there are 3 different user types to choose from:
| Type | Rights |
|---|---|
| admin | ALL |
| machine | WRITE / UPDATE / READ on machines |
| display | READ on machines |

### Rights

The rights can be translated as follows
| right | CRUD equivalent |
|---|---|
| POST | CREATE |
| GET | READ |
| PUT | UPDATE |
| DELETE | DELETE |

### API documentation for the routes

The documentation for the backend API can be found at [https://troppes.github.io/washtrak/](https://troppes.github.io/washtrak/).

### Insomnia

The Insomnia folder in the root directory contains a document for testing all backend routes.

## Token expiry

For this project, the JWT tokens are set to not expire, as the washing machine should update the inputs for a long time. If you need expiry for your project, it can be changed in the `middleware\auth.middleware.js` file.

## Docker

A dockerised version of the backend and documentation can be found at [https://hub.docker.com/r/floreitz/washtrak_frontend](https://hub.docker.com/r/floreitz/washtrak_frontend)
