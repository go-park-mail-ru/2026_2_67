## API

### GET /api/v1/pickup_points

#### Request

#### Response
- `status_code`: 200
- `longitude`, `latitude`: Float — JSON-числа, не строки (без кавычек).
- `Body`:
  ```json
  {
    "pickupPoints": [
      {
        "id": Int,
        "longitude": Float,
        "latitude": Float
      }
    ]
  }
  ```
