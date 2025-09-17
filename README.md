# Meli Challenge: Clasificador de datos sensibles en bases de datos MySQL

Este repositorio contiene una solución para la prueba técnica "Clasificación de base de datos". La aplicación explora instancias MySQL externas, recorre esquemas y tablas, clasifica columnas según reglas configurables (regex) y persiste resultados y un historial de ejecuciones.

# Solución: Clasificación de base de datos

Esta solución explora instancias MySQL externas, recorre esquemas y tablas, y clasifica columnas según reglas configurables (regex) almacenadas en la base de datos. Los resultados y el historial de ejecuciones se persisten para trazabilidad y auditoría.

## Diseño

Se utiliza una base de datos relacional para garantizar integridad referencial, trazabilidad y facilidad de auditoría. El diseño sigue principios SOLID y emplea los patrones Repository y Factory para mantener la solución extensible y testeable. Las reglas de clasificación se gestionan dinámicamente desde la base de datos.

## Cómo ejecutar el proyecto

1. Crear el archivo `.env` en la raíz del proyecto con las variables de entorno necesarias.
2. Ejecutar los servicios con Docker Compose:

```bash
docker-compose up --build -d
```

Una vez levantado, puedes realizar peticiones a los endpoints según el puerto definido en `.env`. Por ejemplo localhost:8000.

## Endpoints principales

### Registrar una base externa

**POST /api/v1/database**

Payload ejemplo:
```json
{
  "host": "meli-challenge-target-db",
  "port": 3309,
  "username": "target_user",
  "password": "target_password"
}
```

Respuesta:
```json
{
  "id": 1
}
```

### Lanzar escaneo

**POST /api/v1/database/scan/:id**

Respuesta:
```json
{
  "scan_id": 1
}
```

### Consultar resultados de escaneo

**GET /api/v1/database/scan/:id**

Respuesta ejemplo:
```json
{
  "database": [
    {
      "schema_name": "target_sample_db",
      "schema_tables": [
        {
          "table_name": "users",
          "columns": [
            {
              "column_name": "created_at",
              "info_type": "N/A"
            },
            {
              "column_name": "first_name",
              "info_type": "FIRST_NAME"
            },
            {
              "column_name": "id",
              "info_type": "N/A"
            },
            {
              "column_name": "ip_address",
              "info_type": "IP_ADDRESS"
            },
            {
              "column_name": "last_name",
              "info_type": "LAST_NAME"
            },
            {
              "column_name": "phone",
              "info_type": "PHONE_NUMBER"
            },
            {
              "column_name": "useremail",
              "info_type": "EMAIL_ADDRESS"
            },
            {
              "column_name": "username",
              "info_type": "USERNAME"
            }
          ]
        }
      ]
    }
  ]
}
```

## Tests

Los tests unitarios están implementados en Testify y cubren la lógica principal del sistema:

1. Validan la creación y actualización de registros en el historial de escaneos (`scan_history`).
2. Verifican la persistencia y agrupamiento de resultados (`scan_results`).
3. Validan que las reglas dinámicas clasifican correctamente columnas, probando casos positivos y negativos para asegurar que los clasificadores construidos desde la base de datos funcionan como se espera.
4. Utilizan mocks (`sqlmock` y repositorios simulados) para probar la lógica sin requerir una base real ni levantar contenedores.
5. Los tests se pueden ejecutar con el siguiente comando:

```bash
go test ./... -v
```

## Modelo de datos

El modelo de datos y las migraciones están definidos en el archivo `init.sql`.

**Resumen del modelo:**

- `external_databases`: almacena las conexiones a bases externas que serán escaneadas (host, puerto, usuario, contraseña).
- `scan_history`: registra cada ejecución de escaneo, con referencia a la base, timestamp y estado (`running`, `success`, `failed`).
- `scan_results`: guarda los resultados detallados de cada escaneo, incluyendo el esquema, tabla, columna y tipo de información detectada.
- `classification_rules`: contiene las reglas de clasificación (regex y tipo), permitiendo que el sistema sea extensible y configurable sin modificar el código.

Las relaciones entre tablas permiten trazabilidad completa: cada resultado está vinculado a un escaneo y cada escaneo a una base registrada.