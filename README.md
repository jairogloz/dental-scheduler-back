# Dental Scheduler Backend

A comprehensive Go backend API for a dental clinic appointment scheduling system using hexagonal/clean architecture.

## Features

- RESTful API for managing clinics, units, doctors, patients, and appointments
- Appointment conflict detection and prevention
- Doctor availability management
- PostgreSQL database with proper indexing and constraints
- Hexagonal/Clean Architecture implementation
- Comprehensive error handling and validation
- CORS support for frontend integration
- Database migrations
- Structured logging
- Health check endpoint

## Architecture

This project follows hexagonal (ports and adapters) architecture with clear separation of concerns:

- **Domain Layer**: Core business entities and logic
- **Application Layer**: Use cases and orchestration
- **Infrastructure Layer**: Database, external services
- **HTTP Layer**: REST API handlers and routing

## Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Docker (optional, for local development)

## Setup

1. Clone the repository
2. Copy environment variables:
   ```bash
   cp .env.example .env
   ```
3. Update database configuration in `.env`
4. Run database migrations:
   ```bash
   make migrate-up
   ```
5. Start the server:
   ```bash
   make run
   ```

## API Endpoints

### Clinics

- `GET /api/v1/clinics` - Get all clinics
- `GET /api/v1/clinics/{id}` - Get specific clinic
- `POST /api/v1/clinics` - Create new clinic
- `PUT /api/v1/clinics/{id}` - Update clinic
- `DELETE /api/v1/clinics/{id}` - Delete clinic

### Units

- `GET /api/v1/units?clinic_id={clinic_id}` - Get units by clinic ID
- `GET /api/v1/units/{id}` - Get specific unit
- `POST /api/v1/units` - Create new unit
- `PUT /api/v1/units/{id}` - Update unit
- `DELETE /api/v1/units/{id}` - Delete unit

### Doctors

- `GET /api/v1/doctors` - Get all doctors
- `GET /api/v1/doctors/{id}` - Get specific doctor
- `GET /api/v1/doctors/{id}/availability?date={date}` - Get doctor's appointments for a specific date
- `POST /api/v1/doctors` - Create new doctor
- `PUT /api/v1/doctors/{id}` - Update doctor
- `DELETE /api/v1/doctors/{id}` - Delete doctor

### Patients

- `GET /api/v1/patients` - Get all patients
- `GET /api/v1/patients/{id}` - Get specific patient
- `POST /api/v1/patients` - Create new patient
- `PUT /api/v1/patients/{id}` - Update patient
- `DELETE /api/v1/patients/{id}` - Delete patient

### Appointments

- `GET /api/v1/appointments` - Get all appointments
- `GET /api/v1/appointments/{id}` - Get specific appointment
- `POST /api/v1/appointments` - Create new appointment (with conflict validation)
- `PUT /api/v1/appointments/{id}` - Update appointment
- `DELETE /api/v1/appointments/{id}` - Delete/cancel appointment
- `GET /api/v1/appointments/upcoming` - Get upcoming appointments

### Doctor Availability

- `GET /api/v1/doctor-availability?doctor_id={id}&date={date}` - Get availability
- `POST /api/v1/doctor-availability` - Create availability rule
- `PUT /api/v1/doctor-availability/{id}` - Update availability
- `DELETE /api/v1/doctor-availability/{id}` - Delete availability

## Development

### Running Tests

```bash
make test
```

### Running with Docker

```bash
docker-compose up
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create name=migration_name
```

## Configuration

Environment variables:

- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL username
- `DB_PASSWORD`: PostgreSQL password
- `DB_NAME`: Database name
- `DB_SSL_MODE`: SSL mode (default: disable)
- `SERVER_PORT`: Server port (default: 8080)
- `SERVER_HOST`: Server host (default: localhost)
- `LOG_LEVEL`: Log level (default: info)
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins

## Project Structure

```
dental-scheduler-backend/
├── cmd/api/               # Application entry point
├── internal/
│   ├── domain/           # Domain layer (entities, ports)
│   ├── app/              # Application layer (use cases, DTOs)
│   ├── infra/            # Infrastructure layer (database, config)
│   └── http/             # HTTP layer (handlers, middleware)
├── pkg/                  # Public packages (utils, errors)
├── scripts/              # Database and utility scripts
└── docker-compose.yml    # Development environment
```
